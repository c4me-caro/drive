package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var keySecret = []byte("RVBIYURVQOB/V%#QB")

func init() {
	godotenv.Load()
	msecret := os.Getenv("SECRET_KEY")
	if msecret != "" {
		keySecret = []byte(msecret)
	}
}

func generateKeys(userId string, randStr string) string {
	p := sha256.New()
	p.Write([]byte(userId))
	userString := hex.EncodeToString(p.Sum(nil))
	message := fmt.Sprintf("%s:%s", userString, randStr)

	h := hmac.New(sha256.New, []byte(keySecret))
	h.Write([]byte(message))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func CreateApiKey(userId string, len int) (string, error) {
	bytes := make([]byte, len)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	randStr := hex.EncodeToString(bytes)
	signature := generateKeys(userId, randStr)

	return fmt.Sprintf("%s:%s", randStr, signature), nil
}

func validateApiKey(key string, userId string) bool {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return false
	}

	randStr := parts[0]
	signature := parts[1]
	expectedSignature := generateKeys(userId, randStr)

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func HandleApiKey(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			userId, err := GetUserIdFromToken(token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				io.WriteString(w, "Error: Unauthorized")
				return
			}

			key := r.URL.Query().Get("key")
			if valid := validateApiKey(key, userId); !valid {
				w.WriteHeader(http.StatusUnauthorized)
				io.WriteString(w, "Error: Unauthorized")
				return
			}

			next.ServeHTTP(w, r)
		})
}
