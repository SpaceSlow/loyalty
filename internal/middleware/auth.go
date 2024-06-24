package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/SpaceSlow/loyalty/internal/config"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func BuildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.ServerConfig.TokenExpiredAt)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(config.ServerConfig.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func getUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(config.ServerConfig.SecretKey), nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil {
		return -1, err
	}

	if !token.Valid {
		return -1, ErrInvalidToken
	}

	return claims.UserID, nil
}

type contextUserID string

const ContextUserIDKey = contextUserID("userID")

func setUserID(ctx *context.Context, userID int) {
	*ctx = context.WithValue(*ctx, ContextUserIDKey, userID)
}

func GetUserID(ctx context.Context) int {
	return ctx.Value(ContextUserIDKey).(int)
}

func WithAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := getUserID(tokenString)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		setUserID(&ctx, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
