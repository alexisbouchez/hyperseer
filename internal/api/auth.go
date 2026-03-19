package api

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"

	"github.com/alexisbouchez/hyperseer/internal/config"
)

func authMiddleware(cfg config.AuthConfig) func(http.Handler) http.Handler {
	if cfg.JWTSecret == "" && cfg.JWKSUrl == "" {
		return func(next http.Handler) http.Handler { return next }
	}

	var keyFunc jwt.Keyfunc
	if cfg.JWKSUrl != "" {
		keyFunc = jwksKeyFunc(cfg.JWKSUrl)
	} else {
		keyFunc = func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if _, err := jwt.Parse(strings.TrimPrefix(auth, "Bearer "), keyFunc); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func jwksKeyFunc(jwksURL string) jwt.Keyfunc {
	var (
		mu   sync.RWMutex
		keys map[string]*rsa.PublicKey
	)

	refresh := func() error {
		fetched, err := fetchJWKS(jwksURL)
		if err != nil {
			return err
		}
		mu.Lock()
		keys = fetched
		mu.Unlock()
		return nil
	}
	_ = refresh()

	return func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		kid, _ := t.Header["kid"].(string)

		mu.RLock()
		key, ok := keys[kid]
		mu.RUnlock()

		if !ok {
			if err := refresh(); err != nil {
				return nil, err
			}
			mu.RLock()
			key, ok = keys[kid]
			mu.RUnlock()
		}
		if !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return key, nil
	}
}

func fetchJWKS(url string) (map[string]*rsa.PublicKey, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	keys := make(map[string]*rsa.PublicKey)
	for _, k := range body.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: int(new(big.Int).SetBytes(eBytes).Int64()),
		}
	}
	return keys, nil
}

// handleAuthConfig returns the auth provider config so the CLI can
// discover it without requiring the user to pass provider flags.
// Intentionally unauthenticated.
func (a *API) handleAuthConfig(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Provider string `json:"provider"`
		URL      string `json:"url"`
		Realm    string `json:"realm,omitempty"`
		ClientID string `json:"client_id,omitempty"`
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{
		Provider: a.cfg.Provider,
		URL:      a.cfg.URL,
		Realm:    a.cfg.Realm,
		ClientID: a.cfg.ClientID,
	})
}
