package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	ConsumerKey    string
	ConsumerSecret string
	APIKey         string
	BaseURL        string
}

func Load() (*Config, error) {
	// Try to load .env file if it exists (for local development)
	_ = godotenv.Load()

	viper.SetEnvPrefix("SPLITWISE")
	viper.AutomaticEnv()

	viper.SetDefault("base_url", "https://secure.splitwise.com/api/v3.0")

	consumerKey := viper.GetString("CONSUMER_KEY")
	consumerSecret := viper.GetString("CONSUMER_SECRET")
	apiKey := viper.GetString("API_KEY")

	if consumerKey == "" {
		return nil, fmt.Errorf("SPLITWISE_CONSUMER_KEY is required")
	}
	if consumerSecret == "" {
		return nil, fmt.Errorf("SPLITWISE_CONSUMER_SECRET is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("SPLITWISE_API_KEY is required")
	}

	return &Config{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		APIKey:         apiKey,
		BaseURL:        viper.GetString("BASE_URL"),
	}, nil
}

func EnsureEnvFile() error {
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		exampleContent := `# Splitwise API Credentials
# Get these from https://secure.splitwise.com/apps
SPLITWISE_CONSUMER_KEY=your_consumer_key_here
SPLITWISE_CONSUMER_SECRET=your_consumer_secret_here
SPLITWISE_API_KEY=your_api_key_here
`
		return os.WriteFile(envFile, []byte(exampleContent), 0644)
	}
	return nil
}
