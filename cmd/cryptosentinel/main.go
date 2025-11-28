package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"teneo-agent-demo1/internal/agent"
	"teneo-agent-demo1/internal/clients"
	"teneo-agent-demo1/internal/services"

	teneo "github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/joho/godotenv"
)

func main() {
	// Untuk lokal: load .env kalau ada.
	// Di Railway, biasanya file .env tidak ada dan ini tidak masalah.
	_ = godotenv.Load()

	ctx := context.Background()

	// --- Ambil dan validasi environment variables penting ---

	privateKey := os.Getenv("PRIVATE_KEY")
	nftTokenID := os.Getenv("NFT_TOKEN_ID")
	ownerAddress := os.Getenv("OWNER_ADDRESS")

	if privateKey == "" {
		log.Fatal("ENV ERROR: PRIVATE_KEY belum di-set. Set di Railway → Variables.")
	}

	if nftTokenID == "" {
		log.Fatal("ENV ERROR: NFT_TOKEN_ID belum di-set. Set di Railway → Variables.")
	}

	if ownerAddress == "" {
		log.Fatal("ENV ERROR: OWNER_ADDRESS belum di-set. Set di Railway → Variables.")
	}

	// (Opsional) kalau kamu punya RATE_LIMIT_PER_MINUTE, bisa diambil di sini juga:
	// rateLimitStr := os.Getenv("RATE_LIMIT_PER_MINUTE")

	// --- HTTP client global ---

	httpClient := &http.Client{
		Timeout: 20 * time.Second,
	}

	// Clients untuk akses API eksternal
	apiClients := clients.NewClients(httpClient)

	// Service untuk alert harga
	alertSvc := services.NewAlertService(apiClients)

	// Agent utama
	appAgent := agent.NewCryptoSentinelAgent(apiClients, alertSvc)

	// Konfigurasi Teneo Agent
	cfg := teneo.DefaultConfig()
	cfg.HealthPort = 9090
	cfg.Name = "CryptoSentinel AI"
	cfg.Description = "CryptoSentinel AI is a crypto market agent that delivers real-time prices, top market overviews, DeFi TVL, CEX gainers comparison, conversions, and price alerts."
	cfg.Capabilities = []string{
		"price_tracking",
		"market_comparison",
		"tvl_monitoring",
		"top_market",
		"gainers_cex",
		"alert_management",
		"chain_tvl",
		"protocol_tvl",
	}


	// Set dari environment
	cfg.PrivateKey = privateKey
	cfg.NFTTokenID = nftTokenID
	cfg.OwnerAddress = ownerAddress

	// Inisialisasi Enhanced Agent
	enhanced, err := teneo.NewEnhancedAgent(&teneo.EnhancedAgentConfig{
		Config:       cfg,
		AgentHandler: appAgent,
	})
	if err != nil {
		log.Fatalf("Failed to create enhanced agent: %v", err)
	}

	// Jalankan background job cek alert
	appAgent.StartBackgroundJobs(ctx)

	log.Println("Starting CryptoSentinel AI...")
	enhanced.Run()
}
