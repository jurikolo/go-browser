package cookie

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// Cookie represents an HTTP cookie with all its attributes
type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Expires  int64  `json:"expires"` // Unix timestamp
	HttpOnly bool   `json:"http_only"`
	Secure   bool   `json:"secure"`
	SameSite string `json:"same_site"`
}

// CookieManager handles cookie operations with ChromeDP integration
type CookieManager struct {
	cookies    []*Cookie
	configDir  string
	encryption bool
	key        []byte
}

// NewCookieManager creates a new CookieManager instance
func NewCookieManager() (*CookieManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "chromedp-browser")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	cm := &CookieManager{
		cookies:   make([]*Cookie, 0),
		configDir: configDir,
		// Encryption is enabled by default
		encryption: true,
		// In a real implementation, this key should be securely generated and stored
		key: []byte("1234567890123456"), // 16 bytes for AES-128
	}

	return cm, nil
}

// LoadCookiesIntoChromedp loads saved cookies into the browser context
func (cm *CookieManager) LoadCookiesIntoChromedp(ctx context.Context) error {
	// Load cookies from disk first
	if err := cm.LoadCookiesFromDisk(); err != nil {
		return fmt.Errorf("failed to load cookies from disk: %w", err)
	}
	
	// Validate that we have valid cookies
	if cm.cookies == nil {
		cm.cookies = make([]*Cookie, 0)
	}

	// Filter out expired cookies and validate required fields
	now := time.Now().Unix()
	validCookies := make([]*Cookie, 0)
	for _, cookie := range cm.cookies {
		// Skip cookies without required fields
		if cookie.Name == "" || cookie.Domain == "" {
			continue
		}
		
		// Skip expired cookies
		if cookie.Expires == 0 || cookie.Expires > now {
			validCookies = append(validCookies, cookie)
		}
	}
	cm.cookies = validCookies

	// Convert our cookies to ChromeDP cookies
	var chromedpCookies []*network.CookieParam
	for _, cookie := range cm.cookies {
		// Decrypt the cookie value if encryption is enabled
		value := cookie.Value
		if cm.encryption && value != "" {
			decryptedValue, err := cm.decrypt(value)
			if err != nil {
				// If decryption fails, log the error and skip this cookie
				// This might happen if the encryption key has changed or data is corrupted
				fmt.Printf("Warning: Failed to decrypt cookie %s: %v\n", cookie.Name, err)
				continue // Skip this cookie
			} else {
				value = decryptedValue
			}
		}

		// Validate required fields
		if cookie.Name == "" {
			return fmt.Errorf("cookie name is required")
		}
		if cookie.Domain == "" {
			return fmt.Errorf("cookie domain is required")
		}

		chromedpCookie := &network.CookieParam{
			Name:     cookie.Name,
			Value:    value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			HTTPOnly: cookie.HttpOnly,
			Secure:   cookie.Secure,
			// Set default values for required fields
			SameSite: network.CookieSameSiteLax,
		}

		// Set expiration time
		if cookie.Expires > 0 {
			expires := cdp.TimeSinceEpoch(time.Unix(cookie.Expires, 0))
			chromedpCookie.Expires = &expires
		}

		// Set SameSite attribute
		switch cookie.SameSite {
		case "Strict":
			chromedpCookie.SameSite = network.CookieSameSiteStrict
		case "Lax":
			chromedpCookie.SameSite = network.CookieSameSiteLax
		case "None":
			chromedpCookie.SameSite = network.CookieSameSiteNone
		default:
			chromedpCookie.SameSite = network.CookieSameSiteLax
		}

		chromedpCookies = append(chromedpCookies, chromedpCookie)
	}

	// Set cookies in ChromeDP
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return network.SetCookies(chromedpCookies).Do(ctx)
	}))
	if err != nil {
		return fmt.Errorf("failed to set cookies in browser: %w", err)
	}
	return nil
}

// ExtractCookiesFromChromedp extracts cookies from the browser context
func (cm *CookieManager) ExtractCookiesFromChromedp(ctx context.Context) error {
	var chromedpCookies []*network.Cookie
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		cookies, err := network.GetCookies().Do(ctx)
		if err != nil {
			return err
		}
		chromedpCookies = cookies
		return nil
	}))
	
	if err != nil {
		return fmt.Errorf("failed to extract cookies from browser: %w", err)
	}

	// Convert ChromeDP cookies to our format
	cm.cookies = make([]*Cookie, len(chromedpCookies))
	for i, chromedpCookie := range chromedpCookies {
		cookie := &Cookie{
			Name:     chromedpCookie.Name,
			Value:    chromedpCookie.Value,
			Domain:   chromedpCookie.Domain,
			Path:     chromedpCookie.Path,
			Expires:  int64(chromedpCookie.Expires),
			HttpOnly: chromedpCookie.HTTPOnly,
			Secure:   chromedpCookie.Secure,
		}

		// Set SameSite attribute
		switch chromedpCookie.SameSite {
		case network.CookieSameSiteStrict:
			cookie.SameSite = "Strict"
		case network.CookieSameSiteNone:
			cookie.SameSite = "None"
		default:
			cookie.SameSite = "Lax"
		}

		// Encrypt the cookie value if encryption is enabled
		if cm.encryption && cookie.Value != "" {
			encryptedValue, err := cm.encrypt(cookie.Value)
			if err != nil {
				return fmt.Errorf("failed to encrypt cookie value: %w", err)
			}
			cookie.Value = encryptedValue
		}

		cm.cookies[i] = cookie
	}

	return nil
}

// SaveCookiesToDisk persists cookies to a JSON file
func (cm *CookieManager) SaveCookiesToDisk() error {
	cookiesFile := filepath.Join(cm.configDir, "cookies.json")

	// Filter out cookies without required fields
	validCookies := make([]*Cookie, 0)
	for _, cookie := range cm.cookies {
		if cookie.Name != "" && cookie.Domain != "" {
			validCookies = append(validCookies, cookie)
		}
	}

	// If no valid cookies, create an empty file
	if len(validCookies) == 0 {
		if err := os.WriteFile(cookiesFile, []byte("[]"), 0644); err != nil {
			return fmt.Errorf("failed to create empty cookies file: %w", err)
		}
		return nil
	}

	data, err := json.MarshalIndent(validCookies, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cookies: %w", err)
	}

	if err := os.WriteFile(cookiesFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cookies to file: %w", err)
	}

	return nil
}

// LoadCookiesFromDisk loads cookies from a JSON file
func (cm *CookieManager) LoadCookiesFromDisk() error {
	cookiesFile := filepath.Join(cm.configDir, "cookies.json")

	// If file doesn't exist, initialize with empty slice
	if _, err := os.Stat(cookiesFile); os.IsNotExist(err) {
		cm.cookies = make([]*Cookie, 0)
		return nil
	}

	data, err := os.ReadFile(cookiesFile)
	if err != nil {
		return fmt.Errorf("failed to read cookies file: %w", err)
	}

	if len(data) == 0 {
		cm.cookies = make([]*Cookie, 0)
		return nil
	}

	var cookies []*Cookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		// If unmarshaling fails, log the error and initialize with empty slice
		fmt.Printf("Warning: Failed to unmarshal cookies, using empty cookie list: %v\n", err)
		cm.cookies = make([]*Cookie, 0)
		return nil
	}
	
	cm.cookies = cookies

	return nil
}

// GetCookies returns all cookies
func (cm *CookieManager) GetCookies() []*Cookie {
	return cm.cookies
}

// GetCookiesForDomain filters cookies by domain
func (cm *CookieManager) GetCookiesForDomain(domain string) []*Cookie {
	var domainCookies []*Cookie
	for _, cookie := range cm.cookies {
		if cookie.Domain == domain || (len(cookie.Domain) > 0 && cookie.Domain[0] == '.' && len(domain) >= len(cookie.Domain)-1 && domain[len(domain)-len(cookie.Domain)+1:] == cookie.Domain[1:]) {
			domainCookies = append(domainCookies, cookie)
		}
	}
	return domainCookies
}

// encrypt encrypts a string using AES and returns base64 encoded string
func (cm *CookieManager) encrypt(plaintext string) (string, error) {
	if !cm.encryption {
		return plaintext, nil
	}

	block, err := aes.NewCipher(cm.key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	// Return base64 encoded string
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts a base64 encoded string using AES
func (cm *CookieManager) decrypt(ciphertext string) (string, error) {
	if !cm.encryption {
		return ciphertext, nil
	}

	// Decode base64 string
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(cm.key)
	if err != nil {
		return "", err
	}

	if len(ciphertextBytes) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertextBytes[:aes.BlockSize]
	ciphertextBytes = ciphertextBytes[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertextBytes, ciphertextBytes)

	return string(ciphertextBytes), nil
}

// EnableEncryption enables or disables cookie value encryption
func (cm *CookieManager) EnableEncryption(enabled bool) {
	cm.encryption = enabled
}

// SetEncryptionKey sets the encryption key
func (cm *CookieManager) SetEncryptionKey(key []byte) error {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return fmt.Errorf("key must be 16, 24, or 32 bytes long")
	}
	cm.key = key
	return nil
}
