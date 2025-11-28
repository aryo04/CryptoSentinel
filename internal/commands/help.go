package commands

func HelpText() string {
	return `📘 CryptoSentinel AI — Command Guide

price [symbol]
  Get real-time price and market data for a coin.
  Example: price btc

compare [symbol1] [symbol2]
  Compare two assets side by side.
  Example: compare btc eth

tvl [protocol]
  Show DeFi protocol TVL (Total Value Locked) from DefiLlama.
  Examples:
    tvl uniswap
    tvl curve-finance

tvlchain [chain]
  Show total TVL for a blockchain network.
  Examples:
    tvlchain ethereum
    tvlchain solana

tvlprotocols [chain]
  Show top DeFi protocols on a chain by TVL.
  Examples:
    tvlprotocols arbitrum
    tvlprotocols avalanche

top [limit]
  List top coins by market cap.
  Examples:
    top
    top 20

digest [daily|now]
  Market summary for top 5 coins.

convert [amount] [from] [to]
  Convert between coins/fiat.

gainers_cex [limit]
  Show top gainers on major CEX (Coinbase, OKX, MEXC, KuCoin, Bitget).

gainers_compare [limit] [cex1] [cex2]
  Compare top gainers between two CEX.

alert [symbol] [condition]
  Create price alert.

alert_list
  Show all alerts.

alert_remove [index]
  Remove an alert.

alert_clear
  Remove all alerts.
  
gainers [limit]
  Show top daily gainers (CoinGecko).
  Examples:
  gainers limit 10

losers [limit]
  Show top daily losers (CoinGecko).
  Examples:
  losers limit 10

feargreed
  Show the global Crypto Fear & Greed Index.

help
  Show this help message again.`
}
