package jwt

import (
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"

	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

// Claims extends jwt.RegisteredClaims with application-specific fields.
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwtv5.RegisteredClaims
}

// Sign creates a signed HS256 JWT with 24h expiry containing the given userID and role.
func Sign(userID uint, role string, secret string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwtv5.NewNumericDate(now),
		},
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Verify parses and validates a JWT string, returning the claims or an error.
// Returns ErrUnauthorized for expired or invalid tokens.
func Verify(tokenStr string, secret string) (*Claims, error) {
	token, err := jwtv5.ParseWithClaims(tokenStr, &Claims{}, func(t *jwtv5.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwtv5.ErrTokenExpired) || errors.Is(err, jwtv5.ErrSignatureInvalid) {
			return nil, apperrors.ErrUnauthorized
		}
		return nil, apperrors.ErrUnauthorized
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apperrors.ErrUnauthorized
	}

	return claims, nil
}
