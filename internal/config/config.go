package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

const (
	appsURL   = "https://secure.splitwise.com/apps"
	devDocsURL = "https://dev.splitwise.com/"
)

type Config struct {
	ConsumerKey    string
	ConsumerSecret string
	APIKey         string
	BaseURL        string
}

type fileConfig struct {
	ConsumerKey    string       `json:"consumer_key"`
	ConsumerSecret string       `json:"consumer_secret"`
	APIKey         string       `json:"api_key"`
	BaseURL        string       `json:"base_url"`
	CurrentUser    *CurrentUser `json:"current_user,omitempty"`
}

// CurrentUser stores the verified current user from the API (used after config).
type CurrentUser struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	DefaultCurrency string `json:"default_currency"`
	Locale          string `json:"locale"`
}

func ConfigPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config", "splitwisecli")
	}
	return filepath.Join(dir, "config.json")
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	consumerKey := os.Getenv("SPLITWISE_CONSUMER_KEY")
	consumerSecret := os.Getenv("SPLITWISE_CONSUMER_SECRET")
	apiKey := os.Getenv("SPLITWISE_API_KEY")
	baseURL := os.Getenv("SPLITWISE_BASE_URL")

	if consumerKey == "" || consumerSecret == "" || apiKey == "" {
		path := ConfigPath()
		data, err := os.ReadFile(path)
		if err == nil {
			var fc fileConfig
			if json.Unmarshal(data, &fc) == nil {
				if consumerKey == "" {
					consumerKey = fc.ConsumerKey
				}
				if consumerSecret == "" {
					consumerSecret = fc.ConsumerSecret
				}
				if apiKey == "" {
					apiKey = fc.APIKey
				}
				if baseURL == "" && fc.BaseURL != "" {
					baseURL = fc.BaseURL
				}
			}
		}
	}

	if consumerKey == "" {
		return nil, fmt.Errorf("SPLITWISE_CONSUMER_KEY is required (set env, or run: splitwisecli config)")
	}
	if consumerSecret == "" {
		return nil, fmt.Errorf("SPLITWISE_CONSUMER_SECRET is required (set env, or run: splitwisecli config)")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("SPLITWISE_API_KEY is required (set env, or run: splitwisecli config)")
	}

	if baseURL == "" {
		baseURL = "https://secure.splitwise.com/api/v3.0"
	}

	return &Config{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		APIKey:         apiKey,
		BaseURL:        baseURL,
	}, nil
}

func Save(cfg *Config) error {
	path := ConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	fc := fileConfig{
		ConsumerKey:    cfg.ConsumerKey,
		ConsumerSecret: cfg.ConsumerSecret,
		APIKey:         cfg.APIKey,
		BaseURL:        cfg.BaseURL,
	}
	data, err := json.MarshalIndent(&fc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// SaveCurrentUser merges the current user into the config file (for verification after setup)
func SaveCurrentUser(id int, name, email, defaultCurrency, locale string) error {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	var fc fileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	fc.CurrentUser = &CurrentUser{
		ID:              id,
		Name:            name,
		Email:           email,
		DefaultCurrency: defaultCurrency,
		Locale:          locale,
	}
	out, err := json.MarshalIndent(&fc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0600)
}

func CredentialSetupInstructions() string {
	return `Getting your Splitwise API credentials
============================================

1. Open the Splitwise Developer Apps page:
   ` + appsURL + `

2. Click "Create new application"

3. Fill in the form:
   - Application name: e.g. "My CLI"
   - Description: optional
   - Accept the API Terms of Use

4. After creating, copy from your app page:
   - Consumer Key
   - Consumer Secret

5. On the app details page, generate your API Key
   (button/link to create a personal API key for your account)

6. Run: splitwisecli config
   and paste your credentials when prompted.

Note: Environment variables (SPLITWISE_CONSUMER_KEY, etc.) override the config file.
`
}

func RunInteractiveSetup() error {
	fmt.Println(CredentialSetupInstructions())
	fmt.Println("Enter your credentials (press Enter after each):")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Consumer Key: ")
	consumerKey, _ := reader.ReadString('\n')
	consumerKey = strings.TrimSpace(consumerKey)

	fmt.Print("Consumer Secret: ")
	consumerSecret, _ := reader.ReadString('\n')
	consumerSecret = strings.TrimSpace(consumerSecret)

	fmt.Print("API Key: ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	if consumerKey == "" || consumerSecret == "" || apiKey == "" {
		return fmt.Errorf("all three credentials are required")
	}

	cfg := &Config{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		APIKey:         apiKey,
		BaseURL:        "https://secure.splitwise.com/api/v3.0",
	}

	if err := Save(cfg); err != nil {
		return err
	}

	fmt.Printf("\nCredentials saved to %s (permissions: 0600)\n", ConfigPath())
	return nil
}

func EnsureEnvFile() error {
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		exampleContent := `# Splitwise API Credentials
# Get these from ` + appsURL + `
SPLITWISE_CONSUMER_KEY=your_consumer_key_here
SPLITWISE_CONSUMER_SECRET=your_consumer_secret_here
SPLITWISE_API_KEY=your_api_key_here
`
		return os.WriteFile(envFile, []byte(exampleContent), 0644)
	}
	return nil
}
