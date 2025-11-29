package commands

func HelpText() string {
	return `📘 CryptoSentinel AI — Command Guide

Core:
  price [symbol]
    Get real-time price and market data for a coin.
    Example:
      price btc

  compare [symbol1] [symbol2]
    Compare two assets side by side.
    Example:
      compare btc eth

  digest [daily|now]
    Market summary for top 5 coins by market cap.
    Examples:
      digest
      digest daily

DeFi & TVL:
  tvl [protocol]
    Show DeFi protocol TVL (Total Value Locked) from DefiLlama.
    Use protocol slug style:
      tvl uniswap
      tvl curve-finance

  tvlchain [chain]
    Show total TVL per chain.
    Example:
      tvlchain ethereum

  tvlprotocols [chain]
    Show top protocols by TVL on a specific chain.
    Example:
      tvlprotocols arbitrum

Market lists:
  top [limit]
    List top coins by market cap (default 10, max 50).
    Examples:
      top
      top 20

  gainers [limit]
    Top daily gainers among large-cap coins (CoinGecko).
    Example:
      gainers 15

  losers [limit]
    Top daily losers among large-cap coins (CoinGecko).
    Example:
      losers 15

CEX gainers & losers:
  gainers_cex [limit]
    Show top gainers on major CEX (Coinbase, Binance, OKX, Bybit, MEXC, KuCoin, Bitget).
    Examples:
      gainers_cex
      gainers_cex 10

  losers_cex [limit]
    Show top losers on the same CEX set.
    Examples:
      losers_cex
      losers_cex 10

  gainers_compare [limit] [cex1] [cex2]
    Compare top gainers between two supported CEX.
    Supported CEX:
      coinbase, binance, okx, bybit, mexc, kucoin, bitget
    Examples:
      gainers_compare 5 binance okx
      gainers_compare 3 mexc kucoin

Conversions:
  convert [amount] [from] [to]
    Convert between coins/fiat using CoinGecko prices.
    Examples:
      convert 1 btc eth
      convert 250 usdt sol
      convert 1000 usd eth

Sentiment:
  feargreed
    Show the Crypto Fear & Greed Index (alternative.me).

News:
  news
    Show latest crypto-related status updates from CoinGecko.

Gas tracker:
  gas [chain]
    Show current gas prices for a chain using Owlracle.
    Supported chains:
      ethereum, arbitrum, base, bsc, polygon, optimism,
      avalanche, fantom, gnosis, celo
    Examples:
      gas ethereum
      gas arbitrum

DEX price:
  dexprice [token] [chain]
    Get on-chain DEX price from Dexscreener.
    Examples:
      dexprice pepe ethereum
      dexprice bonk solana

Portfolio:
  portfolio [address]
    Show crypto portfolio summary for a wallet (via Covalent / DeBank-style APIs).
    Example:
      portfolio 0x1234...abcd

Alerts:
  alert [symbol] [condition]
    Create a price alert with operator >, <, >=, <=.
    Examples:
      alert btc >65000
      alert eth <=3000

  alert_list
    Show all active alerts.

  alert_remove [index]
    Remove an alert by its index (see alert_list).
    Example:
      alert_remove 1

  alert_clear
    Remove all alerts.

Meta:
  help
    Show this help message again.
    Example:
      help`
}
