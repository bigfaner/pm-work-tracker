package jwt

import (
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "pm-work-tracker/backend/internal/pkg/errors"
)

const testSecret = "test-secret-that-is-at-least-32b"

func TestSignAndVerify_ValidToken(t *testing.T) {
	tokenStr, err := Sign(1, "testuser", testSecret)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	claims, err := Verify(tokenStr, testSecret)
	require.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

func TestSignAndVerify_RoundTrip_MultipleUsers(t *testing.T) {
	users := []struct {
		id       uint
		username string
	}{
		{1, "admin"},
		{42, "pm_user"},
		{100, "regular_member"},
	}
	for _, u := range users {
		t.Run(u.username, func(t *testing.T) {
			tokenStr, err := Sign(u.id, u.username, testSecret)
			require.NoError(t, err)

			claims, err := Verify(tokenStr, testSecret)
			require.NoError(t, err)
			assert.Equal(t, u.id, claims.UserID)
			assert.Equal(t, u.username, claims.Username)
		})
	}
}

func TestVerify_ExpiredToken_ReturnsErrUnauthorized(t *testing.T) {
	// Manually create an expired token using -1*time.Second expiry
	claims := &Claims{
		UserID:   1,
		Username: "testuser",
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(-1 * time.Second)),
			IssuedAt:  jwtv5.NewNumericDate(time.Now().Add(-25 * time.Hour)),
		},
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	_, err = Verify(tokenStr, testSecret)
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestVerify_TamperedSignature_ReturnsErrUnauthorized(t *testing.T) {
	tokenStr, err := Sign(1, "testuser", testSecret)
	require.NoError(t, err)

	// Tamper with the token by changing characters at the end
	tampered := tokenStr[:len(tokenStr)-5] + "XXXXX"
	_, err = Verify(tampered, testSecret)
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestVerify_WrongSecret_ReturnsErrUnauthorized(t *testing.T) {
	tokenStr, err := Sign(1, "testuser", testSecret)
	require.NoError(t, err)

	_, err = Verify(tokenStr, "wrong-secret-that-is-also-32-bytes!!")
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestVerify_MalformedToken_ReturnsErrUnauthorized(t *testing.T) {
	_, err := Verify("not-a-valid-token", testSecret)
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
}

func TestSign_Sets24hExpiry(t *testing.T) {
	before := time.Now()
	tokenStr, err := Sign(1, "testuser", testSecret)
	require.NoError(t, err)

	claims, err := Verify(tokenStr, testSecret)
	require.NoError(t, err)

	// Expiry should be approximately 24h from "before"
	expectedExpiry := before.Add(24 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 5*time.Second)
}
