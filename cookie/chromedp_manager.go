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

type CookieManager struct {
	cookies    []*Cookie
	configDir  string
	encryption bool
	key        []byte
}

// Create a new CookieManager instance
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
		cookies:    make([]*Cookie, 0),
		configDir:  configDir,
		encryption: true,
		// TODO: implement secure generation and storage of a key
		// In a real implementation, this key should be securely generated and stored
		key: []byte("1234567890123456"), // 16 bytes for AES-128
	}

	return cm, nil
}

// Loads saved cookies into the browser context
func (cm *CookieManager) LoadCookiesIntoChromedp(ctx context.Context) error {
	if err := cm.LoadCookiesFromDisk(); err != nil {
		return fmt.Errorf("failed to load cookies from disk: %w", err)
	}

	if cm.cookies == nil {
		cm.cookies = make([]*Cookie, 0)
	}

	now := time.Now().Unix()
	validCookies := make([]*Cookie, 0)
	for _, cookie := range cm.cookies {
		if cookie.Name == "" || cookie.Domain == "" {
			continue
		}

		if cookie.Expires == 0 || cookie.Expires > now {
			validCookies = append(validCookies, cookie)
		}
	}
	cm.cookies = validCookies

	// Convert to ChromeDP cookies
	var chromedpCookies []*network.CookieParam
	for _, cookie := range cm.cookies {
		value := cookie.Value
		if cm.encryption && value != "" {
			decryptedValue, err := cm.decrypt(value)
			if err != nil {
				fmt.Printf("Warning: Failed to decrypt cookie %s: %v\n", cookie.Name, err)
				continue
			} else {
				value = decryptedValue
			}
		}

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

	// Now set cookies
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return network.SetCookies(chromedpCookies).Do(ctx)
	}))
	if err != nil {
		return fmt.Errorf("failed to set cookies in browser: %w", err)
	}
	return nil
}

// Extract cookies from the browser context
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

		switch chromedpCookie.SameSite {
		case network.CookieSameSiteStrict:
			cookie.SameSite = "Strict"
		case network.CookieSameSiteNone:
			cookie.SameSite = "None"
		default:
			cookie.SameSite = "Lax"
		}

		// Encrypt the cookie value
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

// Save cookies to a JSON file
func (cm *CookieManager) SaveCookiesToDisk() error {
	cookiesFile := filepath.Join(cm.configDir, "cookies.json")

	validCookies := make([]*Cookie, 0)
	for _, cookie := range cm.cookies {
		if cookie.Name != "" && cookie.Domain != "" {
			validCookies = append(validCookies, cookie)
		}
	}

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

// Load cookies from a JSON file
func (cm *CookieManager) LoadCookiesFromDisk() error {
	cookiesFile := filepath.Join(cm.configDir, "cookies.json")

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
		fmt.Printf("Warning: Failed to unmarshal cookies, using empty cookie list: %v\n", err)
		cm.cookies = make([]*Cookie, 0)
		return nil
	}

	cm.cookies = cookies

	return nil
}

// Return all cookies
func (cm *CookieManager) GetCookies() []*Cookie {
	return cm.cookies
}

// Filter cookies by domain
func (cm *CookieManager) GetCookiesForDomain(domain string) []*Cookie {
	var domainCookies []*Cookie
	for _, cookie := range cm.cookies {
		if cookie.Domain == domain || (len(cookie.Domain) > 0 && cookie.Domain[0] == '.' && len(domain) >= len(cookie.Domain)-1 && domain[len(domain)-len(cookie.Domain)+1:] == cookie.Domain[1:]) {
			domainCookies = append(domainCookies, cookie)
		}
	}
	return domainCookies
}

// Encrypt a string and return base64 encoded string
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

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt a base64 encoded string
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

// Enable or disable cookie value encryption
func (cm *CookieManager) EnableEncryption(enabled bool) {
	cm.encryption = enabled
}

// Set the encryption key
func (cm *CookieManager) SetEncryptionKey(key []byte) error {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return fmt.Errorf("key must be 16, 24, or 32 bytes long")
	}
	cm.key = key
	return nil
}
