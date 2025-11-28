package agent

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// Config represents the configuration for a Teneo agent
type Config struct {
	// Basic agent info
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Image        string   `json:"image"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
	ContactInfo  string   `json:"contact_info"`
	PricingModel string   `json:"pricing_model"`

	// Interface configuration
	InterfaceType  string `json:"interface_type"`
	ResponseFormat string `json:"response_format"`

	// Network configuration
	WebSocketURL     string        `json:"websocket_url"`
	ReconnectEnabled bool          `json:"reconnect_enabled"`
	ReconnectDelay   time.Duration `json:"reconnect_delay"`
	MaxReconnects    int           `json:"max_reconnects"`
	MessageTimeout   time.Duration `json:"message_timeout"`
	PingInterval     time.Duration `json:"ping_interval"`
	HandshakeTimeout time.Duration `json:"handshake_timeout"`

	// Health monitoring
	HealthEnabled bool `json:"health_enabled"`
	HealthPort    int  `json:"health_port"`

	// Authentication
	PrivateKey   string `json:"private_key"`
	OwnerAddress string `json:"owner_address"`
	NFTTokenID   string `json:"nft_token_id"`

	// Room configuration
	Room string `json:"room"`

	// Blockchain configuration
	EthereumRPC        string `json:"ethereum_rpc"`
	NFTContractAddress string `json:"nft_contract_address"`

	// Task processing
	MaxConcurrentTasks int `json:"max_concurrent_tasks"`
	TaskTimeout        int `json:"task_timeout"`
	TaskCheckInterval  int `json:"task_check_interval"`

	// Rate limiting
	RateLimitPerMinute int `json:"rate_limit_per_minute"` // 0 = unlimited
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("agent name is required")
	}
	if c.PrivateKey == "" {
		return fmt.Errorf("private key is required")
	}
	// OwnerAddress is derived from private key, so we don't require it to be set
	return nil
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() error {
	if name := os.Getenv("AGENT_NAME"); name != "" {
		c.Name = name
	}
	if desc := os.Getenv("AGENT_DESCRIPTION"); desc != "" {
		c.Description = desc
	}
	if image := os.Getenv("AGENT_IMAGE"); image != "" {
		c.Image = image
	}
	if version := os.Getenv("AGENT_VERSION"); version != "" {
		c.Version = version
	}
	if caps := os.Getenv("AGENT_CAPABILITIES"); caps != "" {
		c.Capabilities = strings.Split(caps, ",")
	}
	if contact := os.Getenv("AGENT_CONTACT"); contact != "" {
		c.ContactInfo = contact
	}
	if pricing := os.Getenv("AGENT_PRICING"); pricing != "" {
		c.PricingModel = pricing
	}
	if wsURL := os.Getenv("WEBSOCKET_URL"); wsURL != "" {
		c.WebSocketURL = wsURL
	}
	if privateKey := os.Getenv("PRIVATE_KEY"); privateKey != "" {
		c.PrivateKey = privateKey
	}
	if ownerAddr := os.Getenv("OWNER_ADDRESS"); ownerAddr != "" {
		c.OwnerAddress = ownerAddr
	}
	if nftTokenID := os.Getenv("NFT_TOKEN_ID"); nftTokenID != "" {
		c.NFTTokenID = nftTokenID
	}
	if room := os.Getenv("ROOM"); room != "" {
		c.Room = room
	}
	// Also check old env var for backwards compatibility
	if c.Room == "" {
		if roomID := os.Getenv("ROOM_ID"); roomID != "" {
			c.Room = roomID
		}
	}
	if rpc := os.Getenv("ETHEREUM_RPC"); rpc != "" {
		c.EthereumRPC = rpc
	}
	if contract := os.Getenv("NFT_CONTRACT_ADDRESS"); contract != "" {
		c.NFTContractAddress = contract
	}
	if healthPort := os.Getenv("HEALTH_PORT"); healthPort != "" {
		if port, err := strconv.Atoi(healthPort); err == nil {
			c.HealthPort = port
		}
	}
	if rateLimit := os.Getenv("RATE_LIMIT_PER_MINUTE"); rateLimit != "" {
		if limit, err := strconv.Atoi(rateLimit); err == nil {
			c.RateLimitPerMinute = limit
		}
	}
	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Name:               "Teneo Agent",
		Description:        "A Teneo network agent",
		Version:            "1.0.0",
		Capabilities:       []string{"general"},
		InterfaceType:      types.InterfaceTypeNaturalLanguage,
		ResponseFormat:     types.ResponseFormatJSON,
		WebSocketURL:       "wss://backend.developer.chatroom.teneo-protocol.ai/ws",
		ReconnectEnabled:   true,
		ReconnectDelay:     5 * time.Second,
		MaxReconnects:      10,
		MessageTimeout:     30 * time.Second,
		PingInterval:       30 * time.Second,
		HandshakeTimeout:   10 * time.Second,
		HealthEnabled:      true,
		HealthPort:         8080,
		EthereumRPC:        "https://peaq.api.onfinality.io/public",
		NFTContractAddress: "0x811FF962AcBe432344AC974c1111b70847195d3C",
		MaxConcurrentTasks: 5,
		TaskTimeout:        30,
		TaskCheckInterval:  10,
		RateLimitPerMinute: 0, // 0 = unlimited
	}
}
