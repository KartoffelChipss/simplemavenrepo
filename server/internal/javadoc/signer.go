package javadoc

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"
)

// Signer issues tokens that grant read access to one javadoc jar for a limited time.
//
// Javadoc is HTML and JavaScript uploaded by users, so it is shown in a sandboxed iframe with an origin of its own.
// The browser doesn't send the SameSite session cookie from there, the URL carries the access instead.
type Signer struct {
	key []byte
	now func() time.Time
}

func NewSigner(key []byte) *Signer {
	return &Signer{key: key, now: time.Now}
}

// LoadOrCreateKey reads the signing key from the file at p, creating it if it doesn't exist. Keeping the key across
// restarts keeps javadoc pages that are open in browsers working.
func LoadOrCreateKey(p string) ([]byte, error) {
	key, err := os.ReadFile(p)
	if err == nil && len(key) >= 32 {
		return key, nil
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	key = make([]byte, 32)
	rand.Read(key)
	return key, os.WriteFile(p, key, 0o600)
}

// Sign returns a token for the jar at jarPath. Tokens stay the same for a day and are valid for at least another
// day, so the UI's URLs are stable while a page is open.
func (s *Signer) Sign(jarPath string) string {
	day := s.now().UTC().Truncate(24 * time.Hour)
	payload := strconv.FormatInt(day.Add(48*time.Hour).Unix(), 10) + ":" + jarPath
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(s.mac(payload))
}

// Verify returns the jar path of a valid, unexpired token.
func (s *Signer) Verify(token string) (jarPath string, ok bool) {
	encoded, sig, found := strings.Cut(token, ".")
	if !found {
		return "", false
	}
	payload, err1 := base64.RawURLEncoding.DecodeString(encoded)
	mac, err2 := base64.RawURLEncoding.DecodeString(sig)
	if err1 != nil || err2 != nil || !hmac.Equal(mac, s.mac(string(payload))) {
		return "", false
	}
	expiry, jarPath, found := strings.Cut(string(payload), ":")
	exp, err := strconv.ParseInt(expiry, 10, 64)
	if !found || err != nil || s.now().Unix() >= exp {
		return "", false
	}
	return jarPath, true
}

func (s *Signer) mac(payload string) []byte {
	h := hmac.New(sha256.New, s.key)
	h.Write([]byte(payload))
	return h.Sum(nil)
}
