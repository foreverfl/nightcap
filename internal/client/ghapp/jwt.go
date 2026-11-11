package ghapp

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// JWT returns a token that says "I am GitHub App appID", signed with the private key GitHub gave as a .pem at keyPath.
// It lasts about ten minutes. It fails when the file cannot be read or does not hold an RSA private key.
func JWT(appID int64, keyPath string) (string, error) {
	if rest, isHome := strings.CutPrefix(keyPath, "~/"); isHome {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		keyPath = filepath.Join(home, rest)
	}
	pemBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return "", errors.New("ghapp: " + keyPath + " is not a .pem file")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	now := time.Now()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims, err := json.Marshal(map[string]int64{
		"iss": appID,
		"iat": now.Add(-60 * time.Second).Unix(), // GitHub's clock may run a little behind ours
		"exp": now.Add(9 * time.Minute).Unix(),
	})
	if err != nil {
		return "", err
	}
	unsigned := header + "." + base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}
