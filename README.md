# **📘 CryptoSentinel — AI-Powered Crypto Market Intelligence Agent**

**CryptoSentinel** is an intelligent real-time crypto market agent that delivers live price data, CEX/DEX insights, sentiment indicators, volatility analysis, portfolio tracking, TVL analytics, gas monitoring, and automated price alerts — all inside a single fast, modular, AI-driven system.

It is designed for **traders**, **investors**, **DeFi users**, and **on-chain analysts** who need an all-in-one crypto assistant with instant responses powered by public APIs such as **CoinGecko**, **Covalent**, **DefiLlama**, **DexScreener**, **Owlracle**, and centralized exchange data sources.

---

# **🚀 Features**

### **Real-Time Market Intelligence**

* Live prices, charts, and change analysis
* Asset comparison (BTC vs ETH, SOL vs XRP, etc.)
* Trending tokens, gainers & losers
* CEX gainers/losers across Binance, Coinbase, OKX, Bybit, MEXC, KuCoin & Bitget

### **Advanced Analytics**

* Sentiment scoring (market mood index)
* Volatility analysis (risk conditions)
* Trend detection (bullish/bearish structure)

### **DeFi & On-Chain Tools**

* TVL by protocol (Uniswap, Aave, Curve…)
* TVL by chain (Ethereum, Arbitrum, BSC…)
* DEX price scanner (DexScreener)
* Gas tracker (Owlracle)

### **Portfolio Tools**

* Multi-chain portfolio lookups
* Asset values, balances, token metadata (Covalent)

### **Utility Tools**

* News fetcher
* Conversion tool (BTC → ETH, USD → SOL…)
* Price alerts (>, <, >=, <=)

---

# **📁 Folder Structure**

```
.
├── cmd/
│   └── cryptosentinel/
│       └── main.go                # Entry point: bootstraps agent, wiring clients & services
│
├── internal/
│   ├── agent/
│   │   └── agent.go               # CryptoSentinelAgent implementation & background jobs
│   │
│   ├── commands/                  # All chat / slash-style commands
│   │   ├── dispatcher.go          # Router: maps text commands → handler functions
│   │   ├── help.go                # HelpText() – human-readable command guide
│   │   ├── price.go               # price, compare
│   │   ├── digest.go              # digest (now / daily)
│   │   ├── top.go                 # top
│   │   ├── gainers.go             # gainers (CoinGecko)
│   │   ├── losers.go              # losers (CoinGecko)
│   │   ├── gainers_cex.go         # gainers_cex, gainers_compare, losers_cex
│   │   ├── tvl.go                 # tvl [protocol]
│   │   ├── tvl_chain.go           # tvlchain [chain]
│   │   ├── tvl_protocols.go       # tvlprotocols [chain]
│   │   ├── convert.go             # convert [amount] [from] [to]
│   │   ├── gas.go                 # gas [chain] (Owlracle)
│   │   ├── dexprice.go            # dexprice [token] [chain]
│   │   ├── dexmeme.go             # dexmeme [chain] (random meme pairs)
│   │   ├── portfolio.go           # portfolio [address] (Covalent)
│   │   ├── sentiment.go           # sentiment [symbol]
│   │   ├── volatility.go          # volatility [symbol]
│   │   ├── trend.go               # trend [symbol]
│   │   ├── feargreed.go           # feargreed (Crypto Fear & Greed Index)
│   │   ├── news.go                # news (NewsData / CoinGecko status)
│   │   ├── alerts.go              # alert, alert_list, alert_remove, alert_clear
│   │   └── ...                    # future commands live here
│   │
│   ├── clients/                   # External API clients (HTTP wrappers)
│   │   ├── coingecko.go           # prices, markets, coins, IDs
│   │   ├── cex.go                 # CEX gainers/losers (Binance, Coinbase, OKX, Bybit, etc.)
│   │   ├── defillama.go          # protocol & chain TVL
│   │   ├── dexscreener.go         # DEX pair search
│   │   ├── covalent.go            # portfolio & balances
│   │   ├── owlracle.go            # gas data by chain
│   │   ├── newsdata.go            # crypto news feed (if used)
│   │   └── clients.go             # shared HTTP client struct & constructor
│   │
│   ├── services/                  # Domain services (no HTTP here)
│   │   ├── alerts.go              # in-memory alert storage & evaluation
│   │   ├── format.go              # number formatting (F(), IntComma(), etc.)
│   │   └── scheduler.go           # background jobs / tickers (if used)
│   │
│   └── models/                    # Shared data types
│       ├── market.go              # CoinGecko market structs
│       ├── cex.go                 # CexGainer / CexLoser models
│       ├── tvl.go                 # TVL response models
│       ├── portfolio.go           # portfolio / balance models
│       └── alert.go               # alert rule models
│
├── .env.example                   # Example env vars (API keys, config)
├── Dockerfile                     # Docker build definition
├── go.mod                         # Go module definition
├── go.sum                         # Dependency checksums
└── README.md                      # You are here

```
---

# **📜 Command Reference**

Below is the **complete documented list** of all available commands.

---

## **📈 Market Data Commands**

### **💰 `price [symbol]`**

Get real-time price data for any coin.
Example: `price btc`

### **📊 `compare [symbol1] [symbol2]`**

Compare two assets side-by-side.
Example: `compare btc eth`

### **📈 `digest` / `digest daily`**

Quick market summary for top assets.

### **🚀 `gainers [limit]`**

Top market gainers (CoinGecko).
Example: `gainers 20`

### **📉 `losers [limit]`**

Top market losers (CoinGecko).

---

## **🏦 Centralized Exchange Commands**

### **📈 `gainers_cex [limit]`**

Top CEX gainers per exchange
(Coinbase, Binance, OKX, Bybit, MEXC, KuCoin, Bitget)

### **⚔️ `gainers_compare [limit] [cex1] [cex2]`**

Compare the top gainers between two exchanges.
Example: `gainers_compare 5 binance okx`

### **📉 `losers_cex [limit]`**

Top losers per exchange (all supported CEX).

---

## **🧮 Conversion Tools**

### **🔁 `convert [amount] [from] [to]`**

Convert coins or fiat using live rates.
Example: `convert 1 btc eth`

---

## **📉 Market Sentiment & Analytics**

### **😶‍🌫️ `feargreed`**

Crypto Fear & Greed Index.

### **📉 `sentiment [symbol]`**

Market sentiment score (derived from price velocity, volatility & trend).

### **📉 `volatility [symbol]`**

Risk/volatility rating based on recent behavior.

### **📈 `trend [symbol]`**

Trend classification (bullish / neutral / bearish).

---

## **🧩 DeFi Analytics & On-Chain Tools**

### **💧 `tvl [protocol]`**

Get DeFi protocol TVL (Aave, Uniswap, Curve).
Source: DefiLlama.

### **🌐 `tvlchain [chain]`**

TVL for blockchain networks (ETH, Arbitrum, Solana…).

### **📚 `tvlprotocols [chain]`**

Top protocols on a specific chain.

### **💱 `dexprice [token] [chain]`**

DEX price from Dexscreener.

### **⛽ `gas [chain]`**

Gas tracker for EVM chains.
Supports: Ethereum, Arbitrum, Base, BSC, Polygon, Optimism, Avalanche, Fantom, Gnosis, Celo.

---

## **👛 Portfolio Tools**

### **📁 `portfolio [address]`**

Portfolio & balances for EVM addresses
(via Covalent Unified Balances API).

---

## **📰 News**

### **🗞️ `news`**

Latest crypto updates (CoinGecko).

---

## **🔔 Alerts**

### **`alert [symbol] [condition]`**

Create price alert.
Example: `alert btc >65000`

### **`alert_list`**

List active alerts.

### **`alert_remove [index]`**

Remove one alert.

### **`alert_clear`**

Clear all alerts.

---

## **📘 Help Command**

### **❓ `help`**

Show full command list.

---

# **🖼️ Screenshots**

<img width="475" height="564" alt="image" src="https://github.com/user-attachments/assets/aefaf358-f0fc-4d91-b42d-776a611747ff" />

---

# **📚 API Sources**

CryptoSentinel integrates these public data providers:

| API Source                               | Used For                  |
| ---------------------------------------- | ------------------------- |
| **CoinGecko**                            | Prices, market data, news |
| **DefiLlama**                            | TVL per protocol & chain  |
| **Dexscreener**                          | On-chain DEX price lookup |
| **Owlracle**                             | Gas tracker               |
| **Binance/OKX/Bybit/MEXC/KuCoin/Bitget** | CEX spot data             |
| **Covalent**                             | Portfolio, balances       |

---

# **👤 Author**

**Aryo Daffa Khairuddin**
Creator of CryptoSentinel AI
