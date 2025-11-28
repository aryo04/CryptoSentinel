package network

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// HealthStatus represents the health status of a connection
type HealthStatus int32

const (
	// HealthUnknown indicates unknown health status
	HealthUnknown HealthStatus = iota
	// HealthHealthy indicates connection is healthy
	HealthHealthy
	// HealthDegraded indicates connection is degraded but functional
	HealthDegraded
	// HealthUnhealthy indicates connection is unhealthy
	HealthUnhealthy
)

// String returns string representation of health status
func (s HealthStatus) String() string {
	switch s {
	case HealthHealthy:
		return "healthy"
	case HealthDegraded:
		return "degraded"
	case HealthUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// ConnectionMetrics tracks connection health metrics
type ConnectionMetrics struct {
	// Counters
	TotalMessages      int64
	SentMessages       int64
	ReceivedMessages   int64
	FailedMessages     int64
	ReconnectAttempts  int64
	SuccessfulReconnects int64
	
	// Timings
	LastMessageSent    time.Time
	LastMessageReceived time.Time
	LastReconnect      time.Time
	ConnectionEstablished time.Time
	
	// Current state
	IsConnected        bool
	IsAuthenticated    bool
	CurrentLatency     time.Duration
	AverageLatency     time.Duration
	
	// Errors
	ConsecutiveErrors  int
	LastError          error
	LastErrorTime      time.Time
	
	mu sync.RWMutex
}

// HealthMonitor monitors connection health and collects metrics
type HealthMonitor struct {
	metrics         *ConnectionMetrics
	status          int32 // atomic HealthStatus
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	
	// Configuration
	checkInterval   time.Duration
	unhealthyThreshold int
	degradedThreshold  int
	
	// Callbacks
	onStatusChange  func(old, new HealthStatus)
	healthCheckFunc func() error
	
	// Latency tracking
	latencyWindow   []time.Duration
	latencyWindowMu sync.Mutex
	maxLatencySamples int
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(checkInterval time.Duration) *HealthMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &HealthMonitor{
		metrics:            &ConnectionMetrics{},
		status:            int32(HealthUnknown),
		ctx:               ctx,
		cancel:            cancel,
		checkInterval:     checkInterval,
		unhealthyThreshold: 5,  // 5 consecutive errors = unhealthy
		degradedThreshold:  3,  // 3 consecutive errors = degraded
		maxLatencySamples: 100,
		latencyWindow:     make([]time.Duration, 0, 100),
	}
}

// Start begins health monitoring
func (hm *HealthMonitor) Start() {
	hm.wg.Add(1)
	go hm.monitorHealth()
	log.Println("🏥 Health monitor started")
}

// Stop stops health monitoring
func (hm *HealthMonitor) Stop() {
	hm.cancel()
	hm.wg.Wait()
	log.Println("🏥 Health monitor stopped")
}

// SetHealthCheckFunc sets the function used to check health
func (hm *HealthMonitor) SetHealthCheckFunc(fn func() error) {
	hm.healthCheckFunc = fn
}

// SetStatusChangeHandler sets a callback for status changes
func (hm *HealthMonitor) SetStatusChangeHandler(handler func(old, new HealthStatus)) {
	hm.onStatusChange = handler
}

// monitorHealth continuously monitors connection health
func (hm *HealthMonitor) monitorHealth() {
	defer hm.wg.Done()
	
	ticker := time.NewTicker(hm.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-hm.ctx.Done():
			return
			
		case <-ticker.C:
			hm.performHealthCheck()
		}
	}
}

// performHealthCheck performs a health check and updates status
func (hm *HealthMonitor) performHealthCheck() {
	if hm.healthCheckFunc == nil {
		return
	}
	
	err := hm.healthCheckFunc()
	
	hm.metrics.mu.Lock()
	if err != nil {
		hm.metrics.ConsecutiveErrors++
		hm.metrics.LastError = err
		hm.metrics.LastErrorTime = time.Now()
	} else {
		hm.metrics.ConsecutiveErrors = 0
	}
	consecutiveErrors := hm.metrics.ConsecutiveErrors
	hm.metrics.mu.Unlock()
	
	// Determine new status based on consecutive errors
	var newStatus HealthStatus
	if consecutiveErrors == 0 {
		newStatus = HealthHealthy
	} else if consecutiveErrors >= hm.unhealthyThreshold {
		newStatus = HealthUnhealthy
	} else if consecutiveErrors >= hm.degradedThreshold {
		newStatus = HealthDegraded
	} else {
		newStatus = HealthHealthy // Still considered healthy with few errors
	}
	
	hm.updateStatus(newStatus)
}

// updateStatus updates the health status
func (hm *HealthMonitor) updateStatus(newStatus HealthStatus) {
	oldStatus := HealthStatus(atomic.LoadInt32(&hm.status))
	if oldStatus == newStatus {
		return
	}
	
	atomic.StoreInt32(&hm.status, int32(newStatus))
	
	log.Printf("🏥 Health status changed: %s → %s", oldStatus, newStatus)
	
	if hm.onStatusChange != nil {
		go hm.onStatusChange(oldStatus, newStatus)
	}
}

// RecordMessageSent records a sent message
func (hm *HealthMonitor) RecordMessageSent() {
	hm.metrics.mu.Lock()
	defer hm.metrics.mu.Unlock()
	
	hm.metrics.TotalMessages++
	hm.metrics.SentMessages++
	hm.metrics.LastMessageSent = time.Now()
}

// RecordMessageReceived records a received message
func (hm *HealthMonitor) RecordMessageReceived() {
	hm.metrics.mu.Lock()
	defer hm.metrics.mu.Unlock()
	
	hm.metrics.TotalMessages++
	hm.metrics.ReceivedMessages++
	hm.metrics.LastMessageReceived = time.Now()
}

// RecordMessageFailed records a failed message
func (hm *HealthMonitor) RecordMessageFailed() {
	hm.metrics.mu.Lock()
	defer hm.metrics.mu.Unlock()
	
	hm.metrics.FailedMessages++
	hm.metrics.ConsecutiveErrors++
}

// RecordReconnectAttempt records a reconnection attempt
func (hm *HealthMonitor) RecordReconnectAttempt(success bool) {
	hm.metrics.mu.Lock()
	defer hm.metrics.mu.Unlock()
	
	hm.metrics.ReconnectAttempts++
	if success {
		hm.metrics.SuccessfulReconnects++
		hm.metrics.LastReconnect = time.Now()
		hm.metrics.ConsecutiveErrors = 0
	}
}

// RecordConnectionEstablished records when a connection is established
func (hm *HealthMonitor) RecordConnectionEstablished() {
	hm.metrics.mu.Lock()
	defer hm.metrics.mu.Unlock()
	
	hm.metrics.ConnectionEstablished = time.Now()
	hm.metrics.IsConnected = true
	hm.metrics.ConsecutiveErrors = 0
}

// RecordConnectionLost records when a connection is lost
func (hm *HealthMonitor) RecordConnectionLost() {
	hm.metrics.mu.Lock()
	defer hm.metrics.mu.Unlock()
	
	hm.metrics.IsConnected = false
	hm.metrics.IsAuthenticated = false
}

// RecordAuthentication records authentication status
func (hm *HealthMonitor) RecordAuthentication(authenticated bool) {
	hm.metrics.mu.Lock()
	defer hm.metrics.mu.Unlock()
	
	hm.metrics.IsAuthenticated = authenticated
}

// RecordLatency records a latency measurement
func (hm *HealthMonitor) RecordLatency(latency time.Duration) {
	hm.latencyWindowMu.Lock()
	defer hm.latencyWindowMu.Unlock()
	
	// Add to window
	hm.latencyWindow = append(hm.latencyWindow, latency)
	
	// Trim window if too large
	if len(hm.latencyWindow) > hm.maxLatencySamples {
		hm.latencyWindow = hm.latencyWindow[len(hm.latencyWindow)-hm.maxLatencySamples:]
	}
	
	// Calculate average
	var total time.Duration
	for _, l := range hm.latencyWindow {
		total += l
	}
	avgLatency := total / time.Duration(len(hm.latencyWindow))
	
	// Update metrics
	hm.metrics.mu.Lock()
	hm.metrics.CurrentLatency = latency
	hm.metrics.AverageLatency = avgLatency
	hm.metrics.mu.Unlock()
}

// GetStatus returns the current health status
func (hm *HealthMonitor) GetStatus() HealthStatus {
	return HealthStatus(atomic.LoadInt32(&hm.status))
}

// GetMetrics returns a copy of current metrics
func (hm *HealthMonitor) GetMetrics() ConnectionMetrics {
	hm.metrics.mu.RLock()
	defer hm.metrics.mu.RUnlock()
	
	return ConnectionMetrics{
		TotalMessages:        hm.metrics.TotalMessages,
		SentMessages:         hm.metrics.SentMessages,
		ReceivedMessages:     hm.metrics.ReceivedMessages,
		FailedMessages:       hm.metrics.FailedMessages,
		ReconnectAttempts:    hm.metrics.ReconnectAttempts,
		SuccessfulReconnects: hm.metrics.SuccessfulReconnects,
		LastMessageSent:      hm.metrics.LastMessageSent,
		LastMessageReceived:  hm.metrics.LastMessageReceived,
		LastReconnect:        hm.metrics.LastReconnect,
		ConnectionEstablished: hm.metrics.ConnectionEstablished,
		IsConnected:          hm.metrics.IsConnected,
		IsAuthenticated:      hm.metrics.IsAuthenticated,
		CurrentLatency:       hm.metrics.CurrentLatency,
		AverageLatency:       hm.metrics.AverageLatency,
		ConsecutiveErrors:    hm.metrics.ConsecutiveErrors,
		LastError:            hm.metrics.LastError,
		LastErrorTime:        hm.metrics.LastErrorTime,
	}
}

// GetHealthReport returns a formatted health report
func (hm *HealthMonitor) GetHealthReport() string {
	status := hm.GetStatus()
	metrics := hm.GetMetrics()
	
	uptime := time.Duration(0)
	if !metrics.ConnectionEstablished.IsZero() {
		uptime = time.Since(metrics.ConnectionEstablished)
	}
	
	successRate := float64(0)
	if metrics.ReconnectAttempts > 0 {
		successRate = float64(metrics.SuccessfulReconnects) / float64(metrics.ReconnectAttempts) * 100
	}
	
	return fmt.Sprintf(`
Connection Health Report
========================
Status: %s
Connected: %v
Authenticated: %v
Uptime: %v

Messages:
  Total: %d
  Sent: %d
  Received: %d
  Failed: %d

Reconnections:
  Attempts: %d
  Successful: %d
  Success Rate: %.1f%%
  Last Reconnect: %v

Latency:
  Current: %v
  Average: %v

Errors:
  Consecutive: %d
  Last Error: %v
  Last Error Time: %v
`,
		status,
		metrics.IsConnected,
		metrics.IsAuthenticated,
		uptime,
		metrics.TotalMessages,
		metrics.SentMessages,
		metrics.ReceivedMessages,
		metrics.FailedMessages,
		metrics.ReconnectAttempts,
		metrics.SuccessfulReconnects,
		successRate,
		metrics.LastReconnect,
		metrics.CurrentLatency,
		metrics.AverageLatency,
		metrics.ConsecutiveErrors,
		metrics.LastError,
		metrics.LastErrorTime,
	)
}

// IsHealthy returns true if status is healthy
func (hm *HealthMonitor) IsHealthy() bool {
	return hm.GetStatus() == HealthHealthy
}

// IsDegraded returns true if status is degraded
func (hm *HealthMonitor) IsDegraded() bool {
	return hm.GetStatus() == HealthDegraded
}

// IsUnhealthy returns true if status is unhealthy
func (hm *HealthMonitor) IsUnhealthy() bool {
	return hm.GetStatus() == HealthUnhealthy
}