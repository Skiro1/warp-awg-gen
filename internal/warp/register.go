package warp

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/curve25519"
)

const regCacheTTL = 90 * 24 * time.Hour

type RegCache struct {
	SavedAt      time.Time             `json:"saved_at"`
	PrivateKey   []byte                `json:"private_key"`
	PublicKey    []byte                `json:"public_key"`
	Registration *RegistrationResponse `json:"registration"`
}

type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

type AccountInfo struct {
	AccountType string `json:"account_type"`
	WarpPlus    bool   `json:"warp_plus"`
	License     string `json:"license"`
}

type PeerInfo struct {
	PublicKey string `json:"public_key"`
	Endpoint  struct {
		V4 string `json:"v4"`
		V6 string `json:"v6"`
	} `json:"endpoint"`
}

type InterfaceAddresses struct {
	V4 string `json:"v4"`
	V6 string `json:"v6"`
}

type RegistrationResponse struct {
	ID          string               `json:"id"`
	Token       string               `json:"token"`
	Account     AccountInfo          `json:"account"`
	Config      *RegistrationConfig  `json:"config"`
	WarpEnabled bool                 `json:"warp_enabled"`
}

type RegistrationConfig struct {
	ClientID  string       `json:"client_id"`
	Peers     []PeerInfo   `json:"peers"`
	Addresses InterfaceAddresses `json:"-"`
}

func (rc *RegistrationConfig) UnmarshalJSON(data []byte) error {
	var raw struct {
		ClientID  string     `json:"client_id"`
		Peers     []PeerInfo `json:"peers"`
		Interface struct {
			Addresses InterfaceAddresses `json:"addresses"`
		} `json:"interface"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	rc.ClientID = raw.ClientID
	rc.Peers = raw.Peers
	rc.Addresses = raw.Interface.Addresses
	return nil
}

func (rc *RegistrationConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ClientID  string     `json:"client_id"`
		Peers     []PeerInfo `json:"peers"`
		Interface struct {
			Addresses InterfaceAddresses `json:"addresses"`
		} `json:"interface"`
	}{
		ClientID: rc.ClientID,
		Peers:    rc.Peers,
		Interface: struct {
			Addresses InterfaceAddresses `json:"addresses"`
		}{Addresses: rc.Addresses},
	})
}

type apiResponse struct {
	Result  *RegistrationResponse `json:"result"`
	Success bool                  `json:"success"`
}

type registrationRequest struct {
	Key          string `json:"key"`
	InstallID    string `json:"install_id"`
	FCMToken     string `json:"fcm_token"`
	Tos          string `json:"tos"`
	Model        string `json:"model"`
	SerialNumber string `json:"serial_number"`
	Locale       string `json:"locale"`
}

type licenseRequest struct {
	LicenseKey string `json:"license_key"`
}

func GenerateKeyPair() (*KeyPair, error) {
	privateKey := make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	pubKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive public key: %w", err)
	}

	return &KeyPair{PrivateKey: privateKey, PublicKey: pubKey}, nil
}

var apiPrefixes = []string{
	"v0i1909051800",
	"v0a1909051800",
	"v0b1909051800",
	"v0c1909051800",
	"v0d1909051800",
	"v0e1909051800",
	"v0f1909051800",
	"v0g1909051800",
	"v0h1909051800",
	"v0i1604021500",
	"v0a1604021500",
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "0000"
	}
	return hex.EncodeToString(b)
}

func buildRegRequest(kp *KeyPair) *registrationRequest {
	return &registrationRequest{
		Key:          base64.StdEncoding.EncodeToString(kp.PublicKey),
		InstallID:    "",
		FCMToken:     "",
		Tos:          time.Now().UTC().Format(time.RFC3339),
		Model:        "PC",
		SerialNumber: randomHex(8),
		Locale:       "en_US",
	}
}

func doRequest(method, url, authToken string, bodyObj interface{}) (*http.Response, error) {
	var payload []byte
	if bodyObj != nil {
		var err error
		payload, err = json.Marshal(bodyObj)
		if err != nil {
			return nil, fmt.Errorf("marshal error: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "okhttp/3.12.1")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	return client.Do(req)
}

func patchReg(prefix, regID string, bodyObj interface{}, token string) error {
	url := fmt.Sprintf("https://api.cloudflareclient.com/%s/reg/%s", prefix, regID)
	resp, err := doRequest(http.MethodPatch, url, token, bodyObj)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("PATCH returned %d: %s", resp.StatusCode, string(respBody))
}

func Register(kp *KeyPair) (*RegistrationResponse, error) {
	fastClient := &http.Client{Timeout: 5 * time.Second}
	body := buildRegRequest(kp)
	payload, _ := json.Marshal(body)
	var lastErr error

	for _, prefix := range apiPrefixes {
		url := fmt.Sprintf("https://api.cloudflareclient.com/%s/reg", prefix)

		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "okhttp/3.12.1")

		resp, err := fastClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("prefix %s: %w", prefix, err)
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			var wrapped apiResponse
			if err := json.Unmarshal(respBody, &wrapped); err != nil {
				return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(respBody))
			}
			if wrapped.Result == nil {
				return nil, fmt.Errorf("empty result in API response")
			}
			if !wrapped.Result.WarpEnabled {
				_ = patchReg(prefix, wrapped.Result.ID, map[string]interface{}{"warp_enabled": true}, wrapped.Result.Token)
			}
			return wrapped.Result, nil
		}
		if resp.StatusCode == 429 {
			lastErr = fmt.Errorf("rate limited (429)")
			continue
		}
		lastErr = fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil, fmt.Errorf("all API prefixes failed (last: %w)", lastErr)
}

func ApplyLicenseKey(regID, licenseKey, token string) error {
	return patchReg(apiPrefixes[0], regID, licenseRequest{LicenseKey: licenseKey}, token)
}

func SaveRegistrationCache(path string, kp *KeyPair, reg *RegistrationResponse) error {
	cache := RegCache{
		SavedAt:      time.Now(),
		PrivateKey:   kp.PrivateKey,
		PublicKey:    kp.PublicKey,
		Registration: reg,
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	return nil
}

func LoadRegistrationCache(path string) (*KeyPair, *RegistrationResponse, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read cache: %w", err)
	}
	var cache RegCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, nil, fmt.Errorf("parse cache: %w", err)
	}
	if time.Since(cache.SavedAt) > regCacheTTL {
		return nil, nil, fmt.Errorf("cache expired (saved %v)", cache.SavedAt.Format(time.RFC3339))
	}
	return &KeyPair{PrivateKey: cache.PrivateKey, PublicKey: cache.PublicKey}, cache.Registration, nil
}
