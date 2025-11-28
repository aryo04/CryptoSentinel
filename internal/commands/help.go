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
  Use protocol slug style.
  Examples:
    tvl uniswap
    tvl curve-finance

top [limit]
  List top coins by market cap (default 10, max 50).
  Examples:
    top
    top 20

digest [daily|now]
  Market summary for top 5 coins by market cap.
  Examples:
    digest
    digest daily

convert [amount] [from] [to]
  Convert between coins/fiat using CoinGecko prices.
  Examples:
    convert 1 btc eth
    convert 250 usdt sol

gainers_cex [limit]
  Show top gainers on major CEX (Binance, OKX, Bybit, KuCoin, Bitget).
  Examples:
    gainers_cex
    gainers_cex 10

gainers_compare [limit] [cex1] [cex2]
  Compare top gainers between two CEX.
  Examples:
    gainers_compare 5 binance okx
    gainers_compare 3 bybit kucoin

alert [symbol] [condition]
  Create price alert with operator >, <, >=, <=.
  Examples:
    alert btc >65000
    alert eth <=3000

alert_list
  Show all active alerts.
  Example:
    alert_list

alert_remove [index]
  Remove an alert by its index (see alert_list).
  Example:
    alert_remove 0

alert_clear
  Remove all alerts.
  Example:
    alert_clear

help
  Show this help message again.
  Example:
    help`
}
