package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const cookieName = "auth_token"

var secret = []byte("super-secret-key")

func CookieName() string {
	return cookieName
}

func GenerateToken(userID int64) (string, error) {
	data := strconv.FormatInt(userID, 10)

	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write([]byte(data)); err != nil {
		return "", fmt.Errorf("mac write: %w", err)
	}
	sum := mac.Sum(nil)

	token := fmt.Sprintf("%s.%s", data, hex.EncodeToString(sum))
	return base64.URLEncoding.EncodeToString([]byte(token)), nil
}

func ParseToken(token string) (int64, error) {
	raw, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return 0, fmt.Errorf("decode token: %w", err)
	}

	parts := strings.SplitN(string(raw), ".", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("bad token format")
	}

	data, sigHex := parts[0], parts[1]

	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write([]byte(data)); err != nil {
		return 0, fmt.Errorf("mac write: %w", err)
	}
	expected := mac.Sum(nil)

	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return 0, fmt.Errorf("decode sig: %w", err)
	}

	if !hmac.Equal(expected, sig) {
		return 0, fmt.Errorf("bad token signature")
	}

	id, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse id: %w", err)
	}
	return id, nil
}
