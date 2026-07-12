package crypto

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"openriot/logger"
	"openriot/notify"
	"openriot/paths"
)

// Config represents the crypto.toml structure
type Config struct {
	Indicators IndicatorsConfig `toml:"indicators"`
	Pairs      []PairConfig     `toml:"pairs"`
	Display    DisplaySettings  `toml:"display"`
	APIKey     string           `toml:"api_key"`
}

type PairConfig struct {
	Sym   string  `toml:"sym"`
	Coin  string  `toml:"coin"`
	Held  float64 `toml:"held"`
	Entry float64 `toml:"entry"`
}

type IndicatorsConfig struct {
	RSIPeriod  int     `toml:"rsi_period"`
	Oversold   int     `toml:"oversold"`
	Overbought int     `toml:"overbought"`
	BBPeriod   int     `toml:"bb_period"`
	BBStdDev   float64 `toml:"bb_std"`
}

type DisplaySettings struct {
	ShowTotals bool `toml:"show_totals"`
}

// CryptoItem holds processed data for a single cryptocurrency
type CryptoItem struct {
	Sym       string
	CoinID    string
	Held      float64
	Entry     float64
	Price     float64
	PrevPrice float64
	OHLCData  []float64
	Index     int // Preserve config order
	RSI       float64
	BBUpper   float64
	BBLower   float64
	Signal    string // BUY, SELL, HOLD
}

// runCrypto is called from main.go via --crypto flag
func RunCrypto(mode string) error {
	// Load config
	config, err := loadCryptoConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Migrate old cache files and get cache paths
	migrateCacheFiles()
	cacheDir := getCacheDir()
	curFile := filepath.Join(cacheDir, "crypto.json")
	prevFile := filepath.Join(cacheDir, "crypto-prev.json")
	ohlcFile := filepath.Join(cacheDir, "ohlc.json")

	// Fetch prices
	ids := make([]string, 0)
	items := make([]CryptoItem, 0)

	for i, p := range config.Pairs {
		if p.Coin == "" {
			continue
		}
		ids = append(ids, p.Coin)
		items = append(items, CryptoItem{
			Sym:    p.Sym,
			CoinID: p.Coin,
			Held:   p.Held,
			Entry:  p.Entry,
			Index:  i,
		})
	}

	// Load current prices first (becomes "prev")
	curPrices := loadJSON(curFile)

	// Save current as prev BEFORE fetching new prices
	prevData := make(map[string]float64)
	for sym, data := range curPrices {
		if dataMap, ok := data.(map[string]any); ok {
			if usd, ok := dataMap["usd"].(float64); ok {
				prevData[sym] = usd
			}
		}
	}
	if len(prevData) > 0 {
		if data, err := json.Marshal(prevData); err == nil {
			if writeErr := os.WriteFile(prevFile, data, 0600); writeErr != nil {
				logger.Warn(fmt.Sprintf("crypto prev snapshot write: %v", writeErr))
			}
		}
	}

	// Now fetch fresh prices
	fetchPrices(ids, curFile, config.APIKey)

	// Load fresh prices
	curPrices = loadJSON(curFile)

	// Update items with prices
	for i := range items {
		if curData, ok := curPrices[items[i].CoinID].(map[string]any); ok {
			if usd, ok := curData["usd"].(float64); ok {
				items[i].Price = usd
			}
		}
		// Hardcode stablecoin prices (CoinGecko has no fiat coin IDs)
		if items[i].Price == 0 && items[i].Held > 0 {
			switch items[i].Sym {
			case "USD", "USDC", "USDT", "DAI":
				items[i].Price = 1.0
			}
		}
		// Use the prev data we saved above
		if items[i].CoinID != "" {
			if f, ok := prevData[items[i].CoinID]; ok {
				items[i].PrevPrice = f
			}
		}
	}

	// Load OHLC data for RSI
	loadOHLCData(items, ohlcFile, config.APIKey)

	// Apply default indicator values if not set in config
	if config.Indicators.RSIPeriod == 0 {
		config.Indicators.RSIPeriod = 14
	}
	if config.Indicators.Oversold == 0 {
		config.Indicators.Oversold = 30
	}
	if config.Indicators.Overbought == 0 {
		config.Indicators.Overbought = 70
	}
	if config.Indicators.BBPeriod == 0 {
		config.Indicators.BBPeriod = 20
	}
	if config.Indicators.BBStdDev == 0 {
		config.Indicators.BBStdDev = 2.0
	}

	// Calculate indicators for each item
	for i := range items {
		if len(items[i].OHLCData) > config.Indicators.RSIPeriod {
			items[i].RSI = calculateRSI(items[i].OHLCData, config.Indicators.RSIPeriod)
		}
		if len(items[i].OHLCData) > config.Indicators.BBPeriod {
			items[i].BBUpper, items[i].BBLower = calculateBollingerBands(items[i].OHLCData, config.Indicators.BBPeriod, config.Indicators.BBStdDev)
		}
		items[i].Signal = calculateSignal(items[i].RSI, config.Indicators.Oversold, config.Indicators.Overbought)
	}

	// Route to output mode
	mode = strings.ToUpper(mode)

	switch mode {
	case "ROWML":
		return outputROWML(items, config.Display.ShowTotals, curFile, config.Indicators.Oversold)
	case "ROW":
		return outputROW(items)
	case "NOTIFY":
		return outputNotify(items)
	case "NOTIFY_SEND":
		return outputNotifySend(items, config.Display.ShowTotals)
	default:
		// Single symbol mode
		for _, item := range items {
			if strings.ToUpper(item.Sym) == mode {
				return outputSingle(item)
			}
		}
		fmt.Println("--")
		return nil
	}
}

func loadCryptoConfig() (*Config, error) {
	configPath := paths.Join(".config", "crypto.toml")

	// Default config with default pairs - used when config file doesn't exist
	config := &Config{
		Display: DisplaySettings{ShowTotals: false},
		APIKey:  "",
		Pairs: []PairConfig{
			{Sym: "BTC", Coin: "bitcoin", Held: 0, Entry: 0},
			{Sym: "ZEC", Coin: "zcash", Held: 0, Entry: 0},
			{Sym: "XMR", Coin: "monero", Held: 0, Entry: 0},
			{Sym: "LTC", Coin: "litecoin", Held: 0, Entry: 0},
		},
	}

	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		// Return default config if file doesn't exist
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}

	return config, nil
}

func ConfigFileExists() bool {
	_, err := os.Stat(paths.Join(".config", "crypto.toml"))
	return err == nil
}

func getCacheDir() string {
	cacheDir := paths.Join(".cache", "openriot")
	os.MkdirAll(cacheDir, 0700)
	return cacheDir
}

func migrateCacheFiles() {
	oldDir := paths.Join(".cache")
	newDir := paths.Join(".cache", "openriot")
	migrations := map[string]string{
		"openriot-crypto.json":      "crypto.json",
		"openriot-crypto-prev.json": "crypto-prev.json",
		"openriot-ohlc.json":        "ohlc.json",
	}
	for oldName, newName := range migrations {
		oldPath := filepath.Join(oldDir, oldName)
		newPath := filepath.Join(newDir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			if data, err := os.ReadFile(oldPath); err == nil {
				_ = os.WriteFile(newPath, data, 0o600)
				_ = os.Remove(oldPath)
			}
		}
	}
}

func fetchPrices(ids []string, curFile string, apiKey string) {
	if len(ids) == 0 {
		return
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", strings.Join(ids, ","))
	if apiKey != "" {
		url += "&x_cg_demo_api_key=" + apiKey
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "OpenRiot/crypto")

	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var data map[string]any
	if json.NewDecoder(resp.Body).Decode(&data) == nil {
		tmp := curFile + ".tmp"
		f, err := os.Create(tmp)
		if err != nil {
			logger.Warn(fmt.Sprintf("crypto price tmp create: %v", err))
			return
		}
		if err := json.NewEncoder(f).Encode(data); err != nil {
			f.Close()
			os.Remove(tmp)
			logger.Warn(fmt.Sprintf("crypto price encode: %v", err))
			return
		}
		if err := f.Close(); err != nil {
			os.Remove(tmp)
			logger.Warn(fmt.Sprintf("crypto price close: %v", err))
			return
		}
		if err := os.Rename(tmp, curFile); err != nil {
			logger.Warn(fmt.Sprintf("crypto price rename: %v", err))
		}
	}
}

func loadJSON(path string) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		logger.Warn(fmt.Sprintf("crypto cache corrupt: %v", err))
		return nil
	}
	return result
}

func loadOHLCData(items []CryptoItem, ohlcFile string, apiKey string) {
	// Check if cache file exists and is valid
	stat, err := os.Stat(ohlcFile)
	cacheValid := false
	cacheAge := 999999

	if err == nil {
		cacheAge = int(time.Since(stat.ModTime()).Seconds())
		if cacheAge < 1800 && stat.Size() > 10 {
			cacheValid = true
		}
	}

	if cacheValid {
		// Load from cache
		raw := loadJSON(ohlcFile)
		for k, v := range raw {
			if arr, ok := v.([]any); ok {
				prices := make([]float64, len(arr))
				for j, p := range arr {
					if f, ok := p.(float64); ok {
						prices[j] = f
					}
				}
				if len(prices) > 0 {
					for i := range items {
						if items[i].Sym == k {
							items[i].OHLCData = prices
						}
					}
				}
			}
		}
	} else {
		// Refresh OHLC data from API
		ohlcCache := make(map[string][]float64)
		for i := range items {
			prices := fetchOHLC(items[i].CoinID, 180, apiKey)
			if len(prices) > 0 {
				ohlcCache[items[i].Sym] = prices
				items[i].OHLCData = prices
			}
		}
		// Save cache only if we got data
		if len(ohlcCache) > 0 {
			data, err := json.Marshal(ohlcCache)
			if err != nil {
				logger.Warn(fmt.Sprintf("crypto ohlc marshal: %v", err))
				return
			}
			if err := os.WriteFile(ohlcFile, data, 0600); err != nil {
				logger.Warn(fmt.Sprintf("crypto ohlc write: %v", err))
			}
		}
	}
}

func fetchOHLC(coinID string, days int, apiKey string) []float64 {
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s/market_chart?vs_currency=usd&days=%d", coinID, days)
	if apiKey != "" {
		url += "&x_cg_demo_api_key=" + apiKey
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "OpenRiot/crypto")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	if prices, ok := data["prices"].([]any); ok {
		result := make([]float64, len(prices))
		for i, p := range prices {
			if arr, ok := p.([]any); ok && len(arr) > 1 {
				if f, ok := arr[1].(float64); ok {
					result[i] = f
				}
			}
		}
		return result
	}

	return nil
}

// RSI calculation
func calculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 0
	}
	deltas := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		deltas[i-1] = prices[i] - prices[i-1]
	}
	gains := make([]float64, len(deltas))
	losses := make([]float64, len(deltas))
	for i, d := range deltas {
		if d > 0 {
			gains[i] = d
		} else {
			losses[i] = -d
		}
	}
	// Use ALL prior data, not just last 'period' values
	avgGain := 0.0
	avgLoss := 0.0
	for i := range gains {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(len(gains))
	avgLoss /= float64(len(losses))
	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	return math.Round(rsi*10) / 10
}

// Calculate signal based on RSI
func calculateSignal(rsi float64, oversold int, overbought int) string {
	if rsi == 0 {
		return "HOLD"
	}
	if rsi < float64(oversold) {
		return "BUY"
	}
	if rsi > float64(overbought) {
		return "SELL"
	}
	return "HOLD"
}

// Calculate Bollinger Bands
func calculateBollingerBands(prices []float64, period int, stdDev float64) (upper float64, lower float64) {
	if len(prices) < period {
		return 0, 0
	}

	recentPrices := prices[len(prices)-period:]

	sum := 0.0
	for _, p := range recentPrices {
		sum += p
	}
	sma := sum / float64(period)

	variance := 0.0
	for _, p := range recentPrices {
		diff := p - sma
		variance += diff * diff
	}
	std := math.Sqrt(variance / float64(period))

	upper = sma + (stdDev * std)
	lower = sma - (stdDev * std)
	return math.Round(upper*100) / 100, math.Round(lower*100) / 100
}

// calculateTotalPortfolioValue returns the sum of all held coins' current value
func calculateTotalPortfolioValue(items []CryptoItem) float64 {
	total := 0.0
	for _, it := range items {
		if it.Price > 0 && it.Held > 0 {
			total += it.Held * it.Price
		}
	}
	return total
}

// coinIsConcentrated returns true if the coin exceeds 35% of portfolio
func coinIsConcentrated(sym string, items []CryptoItem) bool {
	for _, it := range items {
		if it.Sym == sym && it.Held > 0 && it.Price > 0 {
			coinValue := it.Held * it.Price
			totalValue := calculateTotalPortfolioValue(items)
			if totalValue > 0 && (coinValue/totalValue) > 0.35 {
				return true
			}
		}
	}
	return false
}

// checkConcentration checks if a coin exceeds 50% of portfolio value
func checkConcentration(held, price float64, items []CryptoItem) string {
	if held <= 0 || price <= 0 {
		return ""
	}
	coinValue := held * price
	totalValue := calculateTotalPortfolioValue(items)
	if totalValue > 0 && (coinValue/totalValue) > 0.75 {
		return "⚠"
	}
	return ""
}

// coinPercentOfPortfolio returns the percentage of portfolio a coin represents
func coinPercentOfPortfolio(sym string, items []CryptoItem) float64 {
	coinValue := 0.0
	for _, it := range items {
		if it.Price > 0 && it.Held > 0 && it.Sym == sym {
			coinValue = it.Held * it.Price
			break
		}
	}
	totalValue := calculateTotalPortfolioValue(items)
	if totalValue > 0 {
		return coinValue / totalValue
	}
	return 0
}

// rangeScore computes a momentum score from OHLC data: higher = better buy.
// position in 6-month range * upside potential * recent trend (last 30 days vs older)
func rangeScore(it CryptoItem) float64 {
	if len(it.OHLCData) == 0 || it.Price == 0 {
		return 0
	}
	high := it.OHLCData[0]
	low := it.OHLCData[0]
	for _, p := range it.OHLCData {
		if p > high {
			high = p
		}
		if p < low {
			low = p
		}
	}
	rangeSize := high - low
	if rangeSize <= 0 || it.Price <= low {
		return 0
	}
	position := (it.Price - low) / rangeSize // 0=at low, 1=at high
	upside := high / it.Price                  // >1 if room to run
	// Recent trend: average of last 30 days vs average of older data
	avgRecent := 0.0
	avgOlder := 0.0
	recentDays := 30
	if len(it.OHLCData) > recentDays {
		for i := 0; i < recentDays; i++ {
			avgRecent += it.OHLCData[i]
		}
		avgRecent /= float64(recentDays)
		for i := recentDays; i < len(it.OHLCData); i++ {
			avgOlder += it.OHLCData[i]
		}
		avgOlder /= float64(len(it.OHLCData) - recentDays)
	}
	trend := 1.0
	if avgOlder > 0 {
		trend = avgRecent / avgOlder // >1 if trending up
	}
	base := position * upside * trend
	// RSI modifier: 1.0 at neutral (40-60), bonus for oversold, penalty for overbought
	if it.RSI > 0 {
		if it.RSI > 60 {
			base *= 1.0 - (it.RSI-60)/100
		} else if it.RSI < 40 {
			base *= 1.0 + (40-it.RSI)/100
		}
	}
	return base
}

// rangeRanked returns the best and second-best coin symbols by rangeScore,
// excluding USD stablecoins. Used to determine if the current coin is the
// best performer (it should rotate to the second-best).
func rangeRanked(items []CryptoItem, currentSym string) (best, second string) {
	bestScore := -999.0
	secondScore := -999.0
	for _, it := range items {
		if it.Sym == "USD" || it.Sym == "USDC" {
			continue
		}
		s := rangeScore(it)
		if s > bestScore {
			secondScore = bestScore
			second = best
			bestScore = s
			best = it.Sym
		} else if s > secondScore {
			secondScore = s
			second = it.Sym
		}
	}
	return
}

// findRotationTarget returns the best coin to rotate INTO (highest range momentum score)
func findRotationTarget(items []CryptoItem, oversold int, currentSym string, totalValue, originalInvestment float64) string {
	// If current coin is the best range performer, rotate to second-best
	bestSym, secondSym := rangeRanked(items, currentSym)
	if bestSym == currentSym && secondSym != "" {
		return secondSym
	}

	// Collect candidates and score by range momentum
	type candidate struct {
		sym   string
		score float64
	}
	var candidates []candidate

	for _, it := range items {
		if it.Sym == currentSym || it.Sym == "USD" || it.Sym == "USDC" {
			continue
		}
		if it.Price == 0 || len(it.OHLCData) == 0 {
			continue
		}

		score := rangeScore(it)
		candidates = append(candidates, candidate{sym: it.Sym, score: score})
	}

	bestScore := -999.0
	bestCoin := ""
	for _, c := range candidates {
		if c.score > bestScore {
			bestScore = c.score
			bestCoin = c.sym
		}
	}

	if bestCoin == "" {
		// Below 2.5x original investment: always rotate between cryptos, never USD
		if totalValue < originalInvestment*2.5 {
			bestScore = -999.0
			for _, it := range items {
				if it.Sym == currentSym || it.Sym == "USD" || it.Sym == "USDC" {
					continue
				}
				if it.Price == 0 || len(it.OHLCData) == 0 {
					continue
				}
				s := rangeScore(it)
				if s > bestScore {
					bestScore = s
					bestCoin = it.Sym
				}
			}
		}
	}
	if bestCoin == "" {
		return "USD"
	}
	return bestCoin
}

// calculateSellLimit computes the optimal sell limit (max 20% of entry value)
func calculateSellLimit(sym string, currentPrice, entryPrice, held float64, item CryptoItem, items []CryptoItem, oversold int) string {
	// Skip USD-like stablecoins - you can't sell USD into another token
	upperSym := strings.ToUpper(sym)
	if upperSym == "USD" || upperSym == "USDC" || upperSym == "USDT" || upperSym == "DAI" {
		return ""
	}

	// Calculate max sell value: 20% of entry (entry * held * 0.20)
	maxSellValue := entryPrice * held * 0.20
	if maxSellValue <= 0 || currentPrice <= 0 {
		return ""
	}

	// Calculate max coins we could sell at current price
	maxCoins := maxSellValue / currentPrice

	// Determine best number to sell (20% of holdings, within max)
	var unitsToSell float64
	if held < 1.0 {
		unitsToSell = held * 0.20
		if unitsToSell < 0.01 {
			unitsToSell = held
		}
	} else {
		unitsToSell = math.Min(held*0.20, maxCoins)
		unitsToSell = math.Ceil(unitsToSell*100) / 100
		if unitsToSell < 0.01 {
			unitsToSell = 0.01
		}
	}

	// Calculate target price based on 6-month range and current position
	targetPrice := currentPrice * 1.20
	if len(item.OHLCData) > 0 {
		high := item.OHLCData[0]
		low := item.OHLCData[0]
		for _, p := range item.OHLCData {
			if p > high {
				high = p
			}
			if p < low {
				low = p
			}
		}
		rangeSize := high - low
		if rangeSize > 0 {
			// Where does current price sit in the 6-month range? (0.0 = at low, 1.0 = at high)
			position := (currentPrice - low) / rangeSize
			if position > 0.75 {
				targetPrice = high * 0.98
			} else if position > 0.25 {
				targetPrice = high * 0.90
			} else {
				targetPrice = high * 0.75
			}
		}
	}
	if targetPrice < currentPrice*1.10 {
		targetPrice = currentPrice * 1.10
	}
	if targetPrice > currentPrice*1.50 {
		targetPrice = currentPrice * 1.50
	}

	unitsStr := formatUnits(unitsToSell)
	targetStr := formatPrice(targetPrice)
	totalValue, originalInvestment := 0.0, 0.0
	for _, it := range items {
		if it.Held > 0 && it.Price > 0 {
			totalValue += it.Held * it.Price
		}
		if it.Held > 0 && it.Entry > 0 {
			originalInvestment += it.Held * it.Entry
		}
	}
	rotationTarget := findRotationTarget(items, oversold, sym, totalValue, originalInvestment)

	if currentPrice < entryPrice && entryPrice > 0 {
		return "Hold"
	}

	if rotationTarget == "Hold" {
		return "Hold"
	}

	return fmt.Sprintf("%s @ $%s → %s", unitsStr, targetStr, rotationTarget)
}

// Output formatters matching shell script exactly
// Shell format: f"{held:>9.2f} x" which gives 11 chars (9 for number + space + x)
// In Python f-string, the width applies to the number only, not the added " x"
// So Go must use: fmt.Sprintf("%9.2f", h) + " x" (not "%9.2f x")
func fmtHeld(h float64) string {
	if h > 0 {
		return fmt.Sprintf("%9.2f", h) + " x"
	}
	return ""
}

// fmtHeldClean formats held amount without " x" suffix, right-padded to 7
func fmtHeldClean(h float64) string {
	if h > 0 {
		if h < 1 {
			return fmt.Sprintf("%8s", fmt.Sprintf("%.4f", h))
		}
		return fmt.Sprintf("%8s", fmt.Sprintf("%.2f", h))
	}
	return fmt.Sprintf("%8s", "0.00")
}

// fmtValue formats current value (held \u00d7 price) as whole dollars with commas, right-padded to 7
func fmtValue(held, price float64) string {
	v := held * price
	s := formatIntWithCommas(math.Round(v))
	return fmt.Sprintf("%7s", s)
}

func fmtPrice(price float64) string {
	if math.IsNaN(price) || math.IsInf(price, 0) || price == 0 {
		return "$" + strings.Repeat("-", 10)
	}
	// Format: "$" + 10-char number (with comma separators, right-justified)
	// Shell uses: f"${v:>10,.2f}" which is "$" + 10-char number = 11 chars total
	return "$" + formatNumber(price)
}

func fmtPercent(p float64) string {
	if math.IsNaN(p) || math.IsInf(p, 0) {
		return "  ----%"
	}
	// Shell format: f"  {abs(pct):>5.2f}%" for positive (8 chars)
	// Negative: f"- {abs(pct):>5.2f}%" which is "-" + space + 5-char number = 7 chars total
	if p >= 0 {
		return fmt.Sprintf("  %5.2f%%", p)
	}
	// For negative: use abs value with width specifier to match shell
	return fmt.Sprintf("- %5.2f%%", math.Abs(p))
}

func fmtPercentStable() string {
	// For USD stablecoin: 8 chars "   0.00%"
	return "   0.00%"
}

func fmtAmountStable() string {
	// Shell format: f"     --------" = 8 chars with leading spaces, NO trailing space
	return "     --------"
}

func fmtSignedAmount(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "     --------"
	}
	// Format: " " + %12,.2f for positive, "-" + %12,.2f for negative
	if v >= 0 {
		return fmt.Sprintf(" %12.2f", v)
	}
	return fmt.Sprintf("-%12.2f", -v)
}

func fmtSignedAmountWithSpace(v float64) string {
	// Shell format: f" {abs(v):>12,.2f}" = 13 chars (leading space + 12-char number)
	// NO trailing space - matches shell exactly
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "     -------- "
	}
	// Use abs value for formatting to avoid comma being added to negative sign
	numStr := formatNumberWithWidth(math.Abs(v), 12)
	if v >= 0 {
		return " " + numStr
	}
	return "-" + numStr
}

func outputSingle(item CryptoItem) error {
	if item.Price == 0 {
		fmt.Println("--")
		return nil
	}

	entry := item.Entry
	held := item.Held

	if entry > 0 && held > 0 {
		glAmt := (item.Price - entry) * held
		glPct := ((item.Price - entry) / entry) * 100
		fmt.Printf("%s %s %s %s %s\n",
			item.Sym,
			fmtHeld(held),
			fmtPrice(item.Price),
			fmtPercent(glPct),
			fmtSignedAmount(glAmt))
	} else {
		fmt.Printf("%s %s\n", item.Sym, fmtPrice(item.Price))
	}
	return nil
}

// formatNumber formats with 2 decimal places and comma separators, padded to 10 chars
func formatNumber(v float64) string {
	return formatNumberWithWidth(v, 10)
}

// formatNumberWithWidth formats with 2 decimal places and comma separators, padded to specified width
func formatNumberWithWidth(v float64, width int) string {
	str := fmt.Sprintf("%.2f", v)
	parts := strings.Split(str, ".")
	intPart := parts[0]
	decPart := ""
	if len(parts) > 1 {
		decPart = parts[1]
	}
	var result strings.Builder
	length := len(intPart)
	for i, c := range intPart {
		if i > 0 && (length-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(c)
	}
	intPart = result.String()

	if decPart != "" {
		intPart = intPart + "." + decPart
	}

	if len(intPart) < width {
		intPart = strings.Repeat(" ", width-len(intPart)) + intPart
	}
	return intPart
}

func outputROW(items []CryptoItem) error {
	var parts []string
	for _, item := range items {
		if item.Price == 0 {
			continue
		}
		held := item.Held
		var line string
		if held > 0 && item.Entry > 0 {
			glPct := ((item.Price - item.Entry) / item.Entry) * 100
			line = fmt.Sprintf("%s %s %s %s", item.Sym, fmtHeld(held), fmtPrice(item.Price), fmtPercent(glPct))
		} else {
			line = fmt.Sprintf("%s %s", item.Sym, fmtPrice(item.Price))
		}
		parts = append(parts, line)
	}
	fmt.Println(strings.Join(parts, " • "))
	return nil
}

// sortCryptoItems returns a copy sorted by config order with USD moved to end
func sortCryptoItems(items []CryptoItem) []CryptoItem {
	sorted := make([]CryptoItem, len(items))
	copy(sorted, items)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Sym == "USD" {
			return false
		}
		if sorted[j].Sym == "USD" {
			return true
		}
		return sorted[i].Index < sorted[j].Index
	})
	return sorted
}

// formatPriceArrow returns ▲, ▼, or • based on price change
func formatPriceArrow(item CryptoItem) string {
	if item.Price > 0 && item.PrevPrice > 0 {
		if item.Price > item.PrevPrice {
			return "▲"
		} else if item.Price < item.PrevPrice {
			return "▼"
		}
	}
	return "•"
}

// formatROWMLLine formats a single crypto line for ROWML output as an aligned table
// fmtPriceShort formats price with comma separators, stripping ".00" for whole dollars, right-padded to 8
func fmtPriceShort(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) || v == 0 {
		return fmt.Sprintf("%8s", "")
	}
	str := formatNumberWithWidth(v, 0)
	if strings.HasSuffix(str, ".00") {
		str = str[:len(str)-3]
	}
	return fmt.Sprintf("%8s", str)
}

// formatROWMLLine formats a single crypto line for ROWML output as an aligned table
func formatROWMLLine(item CryptoItem, items []CryptoItem, oversold int) string {
	isUSDStable := item.Sym == "USD" && item.Price >= 0.99 && item.Price <= 1.01

	var heldStr, entryStr, priceStr, valStr, pctStr, portPctStr string
	if item.Held > 0 && item.Entry > 0 && !isUSDStable {
		glPct := ((item.Price - item.Entry) / item.Entry) * 100
		holdStr := fmtHeldClean(item.Held)
		entryStr = fmtPriceShort(item.Entry)
		priceStr = fmtPriceShort(item.Price)
		valStr = fmtValue(item.Held, item.Price)
		pctStr = fmt.Sprintf("%8s", fmt.Sprintf("%+.1f%%", glPct))
		heldStr = holdStr
	} else if isUSDStable {
		heldStr = fmt.Sprintf("%8s", fmt.Sprintf("%.2f", item.Held))
		entryStr = fmt.Sprintf("%8s", fmt.Sprintf("%.2f", item.Entry))
		priceStr = fmt.Sprintf("%8s", fmt.Sprintf("%.2f", item.Price))
		valStr = fmtValue(item.Held, item.Price)
		pctStr = fmt.Sprintf("%8s", "0.0%")
	} else {
		heldStr = fmtHeldClean(item.Held)
		entryStr = ""
		priceStr = fmtPriceShort(item.Price)
		valStr = ""
		pctStr = ""
	}

	// Portfolio percentage
	portPct := coinPercentOfPortfolio(item.Sym, items) * 100
	portPctStr = fmt.Sprintf("%6s", fmt.Sprintf("%.1f%%", portPct))

	sellStr := ""
	if item.Held > 0 && item.Sym != "USD" && item.Sym != "USDC" {
		sellStr = calculateSellLimit(item.Sym, item.Price, item.Entry, item.Held, item, items, oversold)
	}

	return fmt.Sprintf("%-4s %s %s %s %s %s %s %s", item.Sym, heldStr, entryStr, priceStr, valStr, pctStr, portPctStr, sellStr)
}

// calculateTotals returns portfolio totals
func calculateTotals(items []CryptoItem) (heldTotal, gainTotal float64, haveValue, haveGain bool) {
	for _, item := range items {
		if item.Price == 0 {
			continue
		}
		if item.Held > 0 {
			heldTotal += item.Held * item.Price
			haveValue = true
		}
		if item.Held > 0 && item.Entry > 0 {
			gainTotal += (item.Price - item.Entry) * item.Held
			haveGain = true
		}
	}
	return
}

// saveCryptoSnapshot saves current prices as the previous snapshot
func saveCryptoSnapshot(items []CryptoItem, curFile string) {
	prevFile := strings.Replace(curFile, "crypto.json", "crypto-prev.json", 1)
	prevData := make(map[string]float64)
	for _, item := range items {
		if item.Price > 0 {
			prevData[item.Sym] = item.Price
		}
	}
	data, err := json.Marshal(prevData)
	if err != nil {
		logger.Warn(fmt.Sprintf("crypto snapshot marshal: %v", err))
		return
	}
	if err := os.WriteFile(prevFile, data, 0600); err != nil {
		logger.Warn(fmt.Sprintf("crypto snapshot write: %v", err))
	}
}

func outputROWML(items []CryptoItem, showTotals bool, curFile string, oversold int) error {
	lines := []string{
		fmt.Sprintf("%-4s %8s %8s %8s %7s %8s %6s %s", "Coin", "Held", "Entry", "Price", "Value", "Entry", "Port", "Next Step"),
		fmt.Sprintf("%-4s %8s %8s %8s %7s %8s %6s %s", "----", "--------", "--------", "--------", "-------", "--------", "------", "---------"),
	}

	sorted := sortCryptoItems(items)
	displayItems := sorted
	if len(sorted) > 6 {
		displayItems = sorted[:6]
	}

	for _, item := range displayItems {
		lines = append(lines, formatROWMLLine(item, items, oversold))
	}

	heldTotal, gainTotal, haveValue, haveGain := calculateTotals(items)
	if haveGain {
		var gainStr string
		if gainTotal >= 0 {
			gainStr = "+" + formatIntWithCommas(gainTotal)
		} else {
			gainStr = "-" + formatIntWithCommas(-gainTotal)
		}
		gainField := fmt.Sprintf("%8s", gainStr)
		if showTotals && haveValue {
			valField := fmt.Sprintf("%7s", formatIntWithCommas(math.Round(heldTotal)))
			lines = append(lines, fmt.Sprintf("%32s%s %s", "", valField, gainField))
		} else {
			lines = append(lines, fmt.Sprintf("%39s%s", "", gainField))
		}
	}

	fmt.Println(strings.Join(lines, "\n"))
	saveCryptoSnapshot(items, curFile)
	return nil
}

// outputNotify outputs simple format for notifications: "BTC     1.18 x $ 73,237.00 ▲ +10.13%"
func outputNotify(items []CryptoItem) error {
	for _, item := range items {
		arrow := "•"
		if item.Price > 0 && item.PrevPrice > 0 {
			if item.Price > item.PrevPrice {
				arrow = "▲"
			} else if item.Price < item.PrevPrice {
				arrow = "▼"
			}
		}
		pct := ""
		if item.Held > 0 && item.Entry > 0 {
			glPct := ((item.Price - item.Entry) / item.Entry) * 100
			pct = fmt.Sprintf("%.2f%%", glPct)
		}
		fmt.Printf("%-5s %8s x $%12s %s %9s\n", item.Sym, fmt.Sprintf("%.2f", item.Held), formatNumberSimple(item.Price), arrow, pct)
	}
	return nil
}

// getCryptoIcon returns the Nerd Font icon for a given crypto symbol
func getCryptoIcon(sym string) string {
	switch sym {
	case "BTC":
		return ""
	case "XMR":
		return ""
	case "ZEC":
		return "󰰷"
	case "LTC":
		return "󰩡"
	case "USD":
		return ""
	default:
		return "󰦝"
	}
}

// outputNotifySend sends crypto prices as a dunst notification
func outputNotifySend(items []CryptoItem, showTotals bool) error {
	sorted := sortCryptoItems(items)

	var lines []string
	lines = append(lines, "")
	for _, item := range sorted {
		arrow := formatPriceArrow(item)
		pct := ""
		if item.Held > 0 && item.Entry > 0 {
			glPct := ((item.Price - item.Entry) / item.Entry) * 100
			pct = fmt.Sprintf("%.2f%%", glPct)
		}
		lines = append(lines, fmt.Sprintf("%s %-5s %8s x $ %10s %s %7s", getCryptoIcon(item.Sym), item.Sym, fmt.Sprintf("%.2f", item.Held), formatNumberSimple(item.Price), arrow, pct))
	}

	var totalValue, totalCost float64
	for _, item := range sorted {
		totalValue += item.Held * item.Price
		totalCost += item.Held * item.Entry
	}
	totalGain := totalValue - totalCost
	gainArrow := "•"
	if totalGain > 0 {
		gainArrow = "▲"
	} else if totalGain < 0 {
		gainArrow = "▼"
	}
	var gainNumStr string
	if totalGain >= 0 {
		gainNumStr = "+" + formatNumberWithWidth(totalGain, 0)
	} else {
		gainNumStr = "-" + formatNumberWithWidth(-totalGain, 0)
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("󰪚 Total Gains  %s $ %10s", gainArrow, gainNumStr))
	if showTotals {
		heldStr := formatNumberSimple(totalValue)
		lines = append(lines, fmt.Sprintf("󰨖 Total Held     $ %10s", heldStr))
	}

	lines = append(lines, "\n󰳽 Click to Close")
	body := strings.Join(lines, "\n")
	notify.SendNotify("chart", "Crypto Portfolio", body, "normal", 0, 1)
	return nil
}

// formatNumberSimple formats with commas (e.g., 73045 -> 73,045.00)
func formatNumberSimple(v float64) string {
	return formatNumberWithWidth(v, 0)
}

// formatIntWithCommas formats an integer with comma separators (e.g., 55673 -> "55,673")
func formatIntWithCommas(v float64) string {
	intPart := int(math.Round(v))
	str := fmt.Sprintf("%d", intPart)
	var result strings.Builder
	length := len(str)
	for i, c := range str {
		if i > 0 && (length-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(c)
	}
	return result.String()
}

// formatSignedNumber formats with commas and explicit sign (e.g., 1234.56 -> "$ 1,234.56", -1234.56 -> "-$ 1,234.56")
func formatSignedNumber(v float64) string {
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	return sign + "$ " + formatNumberSimple(v)
}
