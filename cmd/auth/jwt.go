package auth

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var secret = []byte("RVBIYURVQOB/V%#QB")
var revokedTokens []string

func init() {
	godotenv.Load()
	msecret := os.Getenv("JWT_SECRET")
	if msecret != "" {
		secret = []byte(msecret)
	}
}

func CreateJWT(userId string) (string, error) {
	claims := &jwt.MapClaims{
		"authorized": true,
		"user_id":    userId,
		"exp":        time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func HandleAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			URLS := []string{"/login", "/logout"}
			skip_urls := strings.Join(URLS, " ")

			if strings.Contains(skip_urls, r.URL.String()) {
				next.ServeHTTP(w, r)
				return
			}

			bearerToken := r.Header["Authorization"][0]
			token := strings.TrimPrefix(bearerToken, "Bearer ")

			for _, revokedToken := range revokedTokens {
				if token == revokedToken {
					w.WriteHeader(http.StatusUnauthorized)
					io.WriteString(w, "Error: Unauthorized")
					return
				}
			}

			_, err := ValidateJWT(token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				io.WriteString(w, "Error: Unauthorized")
				return
			}

			next.ServeHTTP(w, r)
		})
}

func ValidateJWT(token string) (*jwt.Token, error) {
	if token != "" {
		tk, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("sign method not allowed: %v", t.Header["alg"])
			}
			return secret, nil
		})

		if err != nil {
			return nil, err
		}

		if tk.Valid {
			return tk, nil
		}
	}

	return nil, fmt.Errorf("invalid token: %s", token)
}

func InvalidateToken(token string) error {
	_, err := ValidateJWT(token)
	if err != nil {
		return err
	}

	revokedTokens = append(revokedTokens, token)
	return nil
}

func GetUserIdFromToken(token string) (string, error) {
	tk, err := ValidateJWT(token)
	if err != nil {
		return "", err
	}

	claims := tk.Claims.(jwt.MapClaims)
	return claims["user_id"].(string), nil
}
