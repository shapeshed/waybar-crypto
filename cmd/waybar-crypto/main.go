package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/BurntSushi/toml"
	binance "github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
)

var (
	// version and buildDate is set with -ldflags in the Makefile
	Version     string
	BuildDate   string
	configPath  *string
	showVersion *bool
)

// Config structure for TOML
type Config struct {
	APIKey             string `toml:"api_key"`
	SecretKey          string `toml:"secret_key"`
	Ticker             string `toml:"ticker"`
	ShowFundingRate    bool   `toml:"show_funding_rate"`
	ShowOpenInterest   bool   `toml:"show_open_interest"`
	ShowVolumeChange   bool   `toml:"show_volume_change"`
	ShowLongShortRatio bool   `toml:"show_long_short_ratio"`
	ColorPositive      string `toml:"color_positive"`
	ColorNegative      string `toml:"color_negative"`
}

// Get the default config path in XDG_CONFIG_HOME
func getDefaultConfigPath() string {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatalf("Failed to get current user: %v\n", err)
		}
		xdgConfigHome = filepath.Join(usr.HomeDir, ".config")
	}
	return filepath.Join(xdgConfigHome, "waybar-crypto", "config.toml")
}

// Load config from file
func loadConfig(filename string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filename, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Fetch 24h price change statistics
func fetchPriceChangeStats(client *binance.Client, wg *sync.WaitGroup, ticker string, stats *binance.PriceChangeStats) {
	defer wg.Done()
	results, err := client.NewListPriceChangeStatsService().Symbol(ticker).Do(context.Background())
	if err != nil {
		log.Printf("Error fetching price change stats: %v\n", err)
		return
	}
	if len(results) > 0 {
		*stats = *results[0]
	}
}

// Fetch funding rate
func fetchFundingRate(client *futures.Client, wg *sync.WaitGroup, ticker string, fundingRate *float64) {
	defer wg.Done()
	result, err := client.NewFundingRateService().Symbol(ticker).Limit(1).Do(context.Background())
	if err != nil {
		log.Printf("Error fetching funding rate: %v\n", err)
		return
	}
	if len(result) > 0 {
		*fundingRate, _ = strconv.ParseFloat(result[0].FundingRate, 64)
	}
}

// Fetch open interest change
func fetchOpenInterest(client *futures.Client, wg *sync.WaitGroup, ticker string, oiChange *float64) {
	defer wg.Done()
	historicalOI, err := client.NewOpenInterestStatisticsService().
		Symbol(ticker).
		Period("1h").
		Limit(2).
		Do(context.Background())

	if err != nil || len(historicalOI) < 2 {
		log.Printf("Error fetching open interest: %v\n", err)
		return
	}

	currentOIValue, _ := strconv.ParseFloat(historicalOI[1].SumOpenInterest, 64)
	previousOIValue, _ := strconv.ParseFloat(historicalOI[0].SumOpenInterest, 64)

	*oiChange = ((currentOIValue - previousOIValue) / previousOIValue) * 100
}

// Fetch 1-hour volume
func fetchOneHourVolume(client *binance.Client, wg *sync.WaitGroup, ticker string, oneHourVolume *float64) {
	defer wg.Done()
	klines, err := client.NewKlinesService().Symbol(ticker).Interval("1h").Limit(2).Do(context.Background())
	if err != nil {
		log.Printf("Error fetching 1-hour volume: %v\n", err)
		return
	}
	if len(klines) > 1 {
		vol, _ := strconv.ParseFloat(klines[1].Volume, 64)
		*oneHourVolume = vol
	}
}

// Fetch Long-Short Ratio
func fetchLongShortRatio(client *futures.Client, wg *sync.WaitGroup, ticker string, longShortRatio *float64) {
	defer wg.Done()
	result, err := client.NewTopLongShortAccountRatioService().
		Symbol(ticker).
		Period("1h").
		Limit(1).
		Do(context.Background())

	if err != nil || len(result) == 0 {
		log.Printf("Error fetching Long-Short Ratio: %v\n", err)
		return
	}

	*longShortRatio, _ = strconv.ParseFloat(result[0].LongShortRatio, 64)
}

// Determine color coding
func getColor(value float64, config *Config) string {
	if value < 0 {
		return config.ColorNegative
	}
	return config.ColorPositive
}

func main() {
	// Command-line flag for custom config file path
	configPath = flag.String("c", getDefaultConfigPath(), "Path to config file")
	showVersion = flag.Bool("v", false, "Print the version of the program")

	flag.Parse()

	if *showVersion {
		fmt.Printf("Version: %s\nBuild Date: %s\n", Version, BuildDate)
		os.Exit(0)
	}

	// Load configuration
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v\n", err)
	}

	// Initialize Binance clients
	client := binance.NewClient(config.APIKey, config.SecretKey)
	futuresClient := futures.NewClient(config.APIKey, config.SecretKey)

	// Initialize data structures
	var stats binance.PriceChangeStats
	var fundingRate float64
	var openInterest float64
	var oneHourVolume float64
	var longShortRatio float64

	// WaitGroup for concurrent API fetching
	var wg sync.WaitGroup

	wg.Add(1)
	go fetchPriceChangeStats(client, &wg, config.Ticker, &stats)

	if config.ShowFundingRate {
		wg.Add(1)
		go fetchFundingRate(futuresClient, &wg, config.Ticker, &fundingRate)
	}

	if config.ShowOpenInterest {
		wg.Add(1)
		go fetchOpenInterest(futuresClient, &wg, config.Ticker, &openInterest)
	}

	if config.ShowVolumeChange {
		wg.Add(1)
		go fetchOneHourVolume(client, &wg, config.Ticker, &oneHourVolume)
	}

	if config.ShowLongShortRatio {
		wg.Add(1)
		go fetchLongShortRatio(futuresClient, &wg, config.Ticker, &longShortRatio)
	}

	wg.Wait() // Wait for all goroutines to finish

	// Convert values
	lastPrice, _ := strconv.ParseFloat(stats.LastPrice, 64)
	priceChangePercent, _ := strconv.ParseFloat(stats.PriceChangePercent, 64)
	totalVolume, _ := strconv.ParseFloat(stats.Volume, 64)
	volumeChange := (oneHourVolume / (totalVolume / 24)) * 100

	// Construct JSON output
	textOutput := fmt.Sprintf(
		" $%.2f <span color='%s'>%.2f%%</span>",
		lastPrice, getColor(priceChangePercent, config), priceChangePercent,
	)

	tooltipOutput := fmt.Sprintf("Price Change: %.2f%%", priceChangePercent)

	if config.ShowFundingRate {
		textOutput += fmt.Sprintf(" | Funding: <span color='%s'>%.4f%%</span>", getColor(fundingRate, config), fundingRate*100)
		tooltipOutput += fmt.Sprintf(" | Funding Rate: %.4f%%", fundingRate*100)
	}
	if config.ShowOpenInterest {
		textOutput += fmt.Sprintf(" | OI: <span color='%s'>%.2f%%</span>", getColor(openInterest, config), openInterest)
		tooltipOutput += fmt.Sprintf(" | Open Interest Change: %.2f%%", openInterest)
	}
	if config.ShowVolumeChange {
		textOutput += fmt.Sprintf(" | Vol: <span color='%s'>%.2f%%</span>", getColor(volumeChange-100, config), volumeChange-100)
		tooltipOutput += fmt.Sprintf(" | Volume Change: %.2f%%", volumeChange-100)
	}
	if config.ShowLongShortRatio {
		textOutput += fmt.Sprintf(" | LSR: <span color='%s'>%.2f</span>", getColor(longShortRatio-1, config), longShortRatio)
		tooltipOutput += fmt.Sprintf(" | Long-Short Ratio: %.2f", longShortRatio)
	}

	// Construct JSON output
	output := map[string]interface{}{
		"text":    textOutput,
		"tooltip": tooltipOutput,
	}

	// Convert to JSON and print for Waybar
	jsonOutput, _ := json.Marshal(output)
	fmt.Println(string(jsonOutput))
}
