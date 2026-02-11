package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	"github.com/jurikolo/go-browser/internal/utils"
)

// Config represents the browser configuration
type Config struct {
	UserAgent          string        `yaml:"user_agent"`
	Homepage           string        `yaml:"homepage"`
	CookieStorePath    string        `yaml:"cookie_store_path"`
	HistoryStorePath   string        `yaml:"history_store_path"`
	UserDataDir        string        `yaml:"user_data_dir"`
	HeadlessMode       bool          `yaml:"headless_mode"`
	DisableImages      bool          `yaml:"disable_images"`
	NavigationTimeout  time.Duration `yaml:"navigation_timeout"`
	MaxHistorySize     int           `yaml:"max_history_size"`
	ChromeFlags        []string      `yaml:"chrome_flags"`
}

// Default configuration values
const (
	DefaultUserAgent         = "Go-Browser/1.0"
	DefaultHomepage          = "https://jurikolo.name"
	DefaultHeadlessMode      = true
	DefaultDisableImages     = true
	DefaultNavigationTimeout = 60 * time.Second
	DefaultMaxHistorySize    = 1000
)

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "chromedp-browser")
	
	return &Config{
		UserAgent:         DefaultUserAgent,
		Homepage:          DefaultHomepage,
		CookieStorePath:   filepath.Join(configDir, "cookies.json"),
		HistoryStorePath:  filepath.Join(configDir, "history.json"),
		UserDataDir:       configDir,
		HeadlessMode:      DefaultHeadlessMode,
		DisableImages:     DefaultDisableImages,
		NavigationTimeout: DefaultNavigationTimeout,
		MaxHistorySize:    DefaultMaxHistorySize,
		ChromeFlags:       []string{},
	}
}

// Load loads configuration from multiple sources with the following priority:
// 1. Default values (lowest priority)
// 2. YAML file
// 3. Environment variables
// 4. Command-line flags (highest priority)
func Load() (*Config, error) {
	config := NewConfig()
	
	// Load from YAML file
	if err := config.loadFromFile(); err != nil {
		return nil, fmt.Errorf("failed to load config from file: %w", err)
	}
	
	// Load from environment variables
	config.loadFromEnv()
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return config, nil
}

// LoadWithFlags loads configuration including command-line flags
func LoadWithFlags() (*Config, error) {
	config := NewConfig()
	
	// Define command-line flags
	var (
		urlFlag        = flag.String("url", "", "Start with specific URL")
		configFlag     = flag.String("config", "", "Custom config file")
		headlessFlag   = flag.Bool("headless", false, "Force headless mode")
		userDataDirFlag = flag.String("user-data-dir", "", "Custom Chrome profile location")
	)
	
	// Parse command-line flags
	flag.Parse()
	
	// Load from custom config file if specified
	if *configFlag != "" {
		if err := config.loadFromFileAtPath(*configFlag); err != nil {
			return nil, fmt.Errorf("failed to load config from file %s: %w", *configFlag, err)
		}
	} else {
		// Load from default YAML file
		if err := config.loadFromFile(); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}
	
	// Load from environment variables
	config.loadFromEnv()
	
	// Override with command-line flags
	if *urlFlag != "" {
		config.Homepage = *urlFlag
	}
	
	if *headlessFlag {
		config.HeadlessMode = true
	}
	
	if *userDataDirFlag != "" {
		config.UserDataDir = *userDataDirFlag
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return config, nil
}

// loadFromFile loads configuration from the default YAML file
func (c *Config) loadFromFile() error {
	configPath := c.getConfigPath()
	
	// Check if custom config path is provided via environment variable
	if customPath := os.Getenv("CHROMEDP_BROWSER_CONFIG"); customPath != "" {
		configPath = customPath
	}
	
	return c.loadFromFileAtPath(configPath)
}

// loadFromFileAtPath loads configuration from a specific YAML file path
func (c *Config) loadFromFileAtPath(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, use defaults
			return nil
		}
		return err
	}
	
	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return nil
}

// loadFromEnv loads configuration from environment variables
func (c *Config) loadFromEnv() {
	if userAgent := os.Getenv("CHROMEDP_BROWSER_USER_AGENT"); userAgent != "" {
		c.UserAgent = userAgent
	}
	
	if homepage := os.Getenv("CHROMEDP_BROWSER_HOMEPAGE"); homepage != "" {
		c.Homepage = homepage
	}
	
	if cookieStorePath := os.Getenv("CHROMEDP_BROWSER_COOKIE_STORE"); cookieStorePath != "" {
		c.CookieStorePath = cookieStorePath
	}
	
	if historyStorePath := os.Getenv("CHROMEDP_BROWSER_HISTORY_STORE"); historyStorePath != "" {
		c.HistoryStorePath = historyStorePath
	}
	
	if userDataDir := os.Getenv("CHROMEDP_BROWSER_USER_DATA_DIR"); userDataDir != "" {
		c.UserDataDir = userDataDir
	}
	
	if headless := os.Getenv("CHROMEDP_BROWSER_HEADLESS"); headless != "" {
		c.HeadlessMode = strings.ToLower(headless) == "true"
	}
	
	if disableImages := os.Getenv("CHROMEDP_BROWSER_DISABLE_IMAGES"); disableImages != "" {
		c.DisableImages = strings.ToLower(disableImages) == "true"
	}
	
	if timeout := os.Getenv("CHROMEDP_BROWSER_NAVIGATION_TIMEOUT"); timeout != "" {
		if dur, err := time.ParseDuration(timeout); err == nil {
			c.NavigationTimeout = dur
		}
	}
	
	if maxHistorySize := os.Getenv("CHROMEDP_BROWSER_MAX_HISTORY_SIZE"); maxHistorySize != "" {
		if size, err := fmt.Sscanf(maxHistorySize, "%d", &c.MaxHistorySize); err != nil || size != 1 {
			// Use default if parsing fails
		}
	}
}

func (c *Config) SetUrlPrefix() {
	c.Homepage = utils.SetUrlPrefix(c.Homepage)
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	c.SetUrlPrefix()
	// Validate homepage URL format
	if c.Homepage != "" && !strings.HasPrefix(c.Homepage, "http://") && !strings.HasPrefix(c.Homepage, "https://") {
		return fmt.Errorf("homepage must be a valid URL starting with http:// or https://")
	}
	
	// Validate timeout values
	if c.NavigationTimeout <= 0 {
		return fmt.Errorf("navigation timeout must be positive")
	}
	
	if c.NavigationTimeout > 5*time.Minute {
		return fmt.Errorf("navigation timeout is too long (max 5 minutes)")
	}
	
	// Validate max history size
	if c.MaxHistorySize <= 0 {
		return fmt.Errorf("max history size must be positive")
	}
	
	if c.MaxHistorySize > 100000 {
		return fmt.Errorf("max history size is too large (max 100000)")
	}
	
	// Ensure paths exist or create them
	if err := c.EnsurePaths(); err != nil {
		return fmt.Errorf("failed to ensure paths: %w", err)
	}
	
	return nil
}

// EnsurePaths ensures that required directories exist
func (c *Config) EnsurePaths() error {
	// Ensure user data directory exists
	if err := os.MkdirAll(c.UserDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create user data directory: %w", err)
	}
	
	// Ensure cookie store directory exists
	if cookieDir := filepath.Dir(c.CookieStorePath); cookieDir != "." && cookieDir != "/" {
		if err := os.MkdirAll(cookieDir, 0755); err != nil {
			return fmt.Errorf("failed to create cookie store directory: %w", err)
		}
	}
	
	// Ensure history store directory exists
	if historyDir := filepath.Dir(c.HistoryStorePath); historyDir != "." && historyDir != "/" {
		if err := os.MkdirAll(historyDir, 0755); err != nil {
			return fmt.Errorf("failed to create history store directory: %w", err)
		}
	}
	
	return nil
}

// Save saves the current configuration to the YAML file
func (c *Config) Save() error {
	configPath := c.getConfigPath()
	
	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// getConfigPath returns the default configuration file path
func (c *Config) getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "chromedp-browser", "config.yaml")
}
