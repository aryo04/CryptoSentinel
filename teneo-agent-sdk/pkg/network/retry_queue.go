package network

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// RetryPolicy defines how messages should be retried
type RetryPolicy struct {
	MaxRetries     int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	BackoffFactor  float64
	RetryableError func(error) bool
}

// DefaultRetryPolicy returns a default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableError: func(err error) bool {
			// By default, all errors are retriable
			// Can be customized to check for specific error types
			return true
		},
	}
}

// RetryableMessage represents a message that can be retried
type RetryableMessage struct {
	Message     *types.Message
	RetryCount  int
	LastAttempt time.Time
	NextRetry   time.Time
	Error       error
}

// MessageRetryQueue manages failed messages for retry
type MessageRetryQueue struct {
	queue      []*RetryableMessage
	policy     *RetryPolicy
	sendFunc   func(*types.Message) error
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	processing bool
	metrics    *RetryMetrics
}

// RetryMetrics tracks retry queue statistics
type RetryMetrics struct {
	TotalRetries      int64
	SuccessfulRetries int64
	FailedRetries     int64
	DroppedMessages   int64
	CurrentQueueSize  int
	mu                sync.RWMutex
}

// NewMessageRetryQueue creates a new message retry queue
func NewMessageRetryQueue(policy *RetryPolicy, sendFunc func(*types.Message) error) *MessageRetryQueue {
	if policy == nil {
		policy = DefaultRetryPolicy()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &MessageRetryQueue{
		queue:    make([]*RetryableMessage, 0),
		policy:   policy,
		sendFunc: sendFunc,
		ctx:      ctx,
		cancel:   cancel,
		metrics:  &RetryMetrics{},
	}
}

// Start begins processing the retry queue
func (q *MessageRetryQueue) Start() {
	q.mu.Lock()
	if q.processing {
		q.mu.Unlock()
		return
	}
	q.processing = true
	q.mu.Unlock()

	q.wg.Add(1)
	go q.processQueue()

	log.Println("📮 Message retry queue started")
}

// Stop stops processing the retry queue
func (q *MessageRetryQueue) Stop() {
	q.mu.Lock()
	if !q.processing {
		q.mu.Unlock()
		return
	}
	q.processing = false
	q.mu.Unlock()

	q.cancel()
	q.wg.Wait()

	log.Printf("📮 Message retry queue stopped. Dropped %d messages", len(q.queue))
}

// Enqueue adds a failed message to the retry queue
func (q *MessageRetryQueue) Enqueue(msg *types.Message, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if error is retriable
	if !q.policy.RetryableError(err) {
		log.Printf("⚠️ Message not retriable: %v", err)
		q.updateMetrics(func(m *RetryMetrics) {
			m.DroppedMessages++
		})
		return
	}

	retryMsg := &RetryableMessage{
		Message:     msg,
		RetryCount:  0,
		LastAttempt: time.Now(),
		NextRetry:   time.Now().Add(q.policy.InitialDelay),
		Error:       err,
	}

	q.queue = append(q.queue, retryMsg)
	q.updateMetrics(func(m *RetryMetrics) {
		m.CurrentQueueSize = len(q.queue)
	})

	log.Printf("📮 Message queued for retry (queue size: %d)", len(q.queue))
}

// processQueue continuously processes messages in the retry queue
func (q *MessageRetryQueue) processQueue() {
	defer q.wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-q.ctx.Done():
			return

		case <-ticker.C:
			q.processReadyMessages()
		}
	}
}

// processReadyMessages processes messages that are ready for retry
func (q *MessageRetryQueue) processReadyMessages() {
	q.mu.Lock()

	now := time.Now()
	readyMessages := make([]*RetryableMessage, 0)
	remainingMessages := make([]*RetryableMessage, 0)

	// Separate ready messages from those still waiting
	for _, msg := range q.queue {
		if now.After(msg.NextRetry) {
			readyMessages = append(readyMessages, msg)
		} else {
			remainingMessages = append(remainingMessages, msg)
		}
	}

	q.queue = remainingMessages
	q.updateMetricsLocked(func(m *RetryMetrics) {
		m.CurrentQueueSize = len(q.queue)
	})
	q.mu.Unlock()

	// Process ready messages without holding the lock
	for _, retryMsg := range readyMessages {
		q.retryMessage(retryMsg)
	}
}

// retryMessage attempts to retry a single message
func (q *MessageRetryQueue) retryMessage(retryMsg *RetryableMessage) {
	retryMsg.RetryCount++
	retryMsg.LastAttempt = time.Now()

	log.Printf("🔄 Retrying message (attempt %d/%d)", retryMsg.RetryCount, q.policy.MaxRetries)

	// Attempt to send the message
	err := q.sendFunc(retryMsg.Message)

	if err == nil {
		// Success!
		log.Printf("✅ Message retry successful after %d attempts", retryMsg.RetryCount)
		q.updateMetrics(func(m *RetryMetrics) {
			m.SuccessfulRetries++
			m.TotalRetries++
		})
		return
	}

	// Failed again
	retryMsg.Error = err
	q.updateMetrics(func(m *RetryMetrics) {
		m.TotalRetries++
	})

	// Check if we should retry again
	if retryMsg.RetryCount >= q.policy.MaxRetries {
		log.Printf("❌ Message dropped after %d retries: %v", retryMsg.RetryCount, err)
		q.updateMetrics(func(m *RetryMetrics) {
			m.FailedRetries++
			m.DroppedMessages++
		})
		return
	}

	// Calculate next retry time with exponential backoff
	delay := q.calculateBackoff(retryMsg.RetryCount)
	retryMsg.NextRetry = time.Now().Add(delay)

	// Re-queue the message
	q.mu.Lock()
	q.queue = append(q.queue, retryMsg)
	q.updateMetricsLocked(func(m *RetryMetrics) {
		m.CurrentQueueSize = len(q.queue)
	})
	q.mu.Unlock()

	log.Printf("📮 Message re-queued for retry in %v", delay)
}

// calculateBackoff calculates the backoff delay for a retry attempt
func (q *MessageRetryQueue) calculateBackoff(retryCount int) time.Duration {
	delay := float64(q.policy.InitialDelay)

	for i := 1; i < retryCount; i++ {
		delay *= q.policy.BackoffFactor
	}

	if time.Duration(delay) > q.policy.MaxDelay {
		return q.policy.MaxDelay
	}

	return time.Duration(delay)
}

// GetMetrics returns current retry queue metrics
func (q *MessageRetryQueue) GetMetrics() RetryMetrics {
	q.metrics.mu.RLock()
	defer q.metrics.mu.RUnlock()

	return RetryMetrics{
		TotalRetries:      q.metrics.TotalRetries,
		SuccessfulRetries: q.metrics.SuccessfulRetries,
		FailedRetries:     q.metrics.FailedRetries,
		DroppedMessages:   q.metrics.DroppedMessages,
		CurrentQueueSize:  q.metrics.CurrentQueueSize,
	}
}

// updateMetrics safely updates metrics
func (q *MessageRetryQueue) updateMetrics(update func(*RetryMetrics)) {
	q.metrics.mu.Lock()
	defer q.metrics.mu.Unlock()
	update(q.metrics)
}

// updateMetricsLocked updates metrics when already holding queue lock
func (q *MessageRetryQueue) updateMetricsLocked(update func(*RetryMetrics)) {
	q.metrics.mu.Lock()
	defer q.metrics.mu.Unlock()
	update(q.metrics)
}

// GetQueueSize returns the current size of the retry queue
func (q *MessageRetryQueue) GetQueueSize() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue)
}

// Clear removes all messages from the retry queue
func (q *MessageRetryQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	dropped := len(q.queue)
	q.queue = make([]*RetryableMessage, 0)

	q.updateMetricsLocked(func(m *RetryMetrics) {
		m.DroppedMessages += int64(dropped)
		m.CurrentQueueSize = 0
	})

	log.Printf("📮 Retry queue cleared. Dropped %d messages", dropped)
}

// String returns a string representation of metrics
func (m *RetryMetrics) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf(
		"RetryMetrics{Total: %d, Success: %d, Failed: %d, Dropped: %d, QueueSize: %d}",
		m.TotalRetries, m.SuccessfulRetries, m.FailedRetries,
		m.DroppedMessages, m.CurrentQueueSize,
	)
}
