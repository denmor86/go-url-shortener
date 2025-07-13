package helpers

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildJWT(t *testing.T) {
	// Определяем тестовые случаи
	testCases := []struct {
		name      string
		userID    string
		secret    []byte
		wantError bool
	}{
		{
			name:      "Successful test #1 (good)",
			userID:    "mda",
			secret:    []byte("valid-secret-key"),
			wantError: false,
		},
		{
			name:      "Empty user #2 (bad)",
			userID:    "",
			secret:    []byte("valid-secret-key"),
			wantError: false, // пустой userID может быть допустимым в некоторых случаях
		},
		{
			name:      "Empty secret #3 (bad)",
			userID:    "mda",
			secret:    []byte(""),
			wantError: true, // пустой секретный ключ должен вызывать ошибку
		},
		{
			name:      "Nil secret #4 (bad)",
			userID:    "mda",
			secret:    nil,
			wantError: true, // nil секретный ключ должен вызывать ошибку
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Вызываем тестируемую функцию
			tokenString, err := BuildJWT(tc.userID, tc.secret)

			// Проверяем ожидаемую ошибку
			if tc.wantError {
				assert.Error(t, err, "expected error but got none")
				assert.Empty(t, tokenString, "token should be empty when error occurs")
				return
			}

			// Если ошибка не ожидается
			require.NoError(t, err, "unexpected error")
			assert.NotEmpty(t, tokenString, "token should not be empty")

			// Парсим токен для проверки его содержимого
			token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				return tc.secret, nil
			})

			require.NoError(t, err, "failed to parse token")
			assert.True(t, token.Valid, "token should be valid")

			// Проверяем claims
			claims, ok := token.Claims.(*JWTClaims)
			require.True(t, ok, "invalid claims type")

			assert.Equal(t, tc.userID, claims.UserID, "user ID in claims doesn't match")
			assert.WithinDuration(t, time.Now().Add(JWTExpire), claims.ExpiresAt.Time, time.Second, "expiration time is not correct")
		})
	}
}

func TestParseJWT(t *testing.T) {
	// Создадим валидный токен для тестов
	validUserID := "mda"
	validSecret := []byte("valid-secret-key")
	validToken, err := BuildJWT(validUserID, validSecret)
	require.NoError(t, err, "failed to create valid test token")

	testCases := []struct {
		name        string
		tokenString string
		secret      []byte
		wantError   bool
		errorText   string
	}{
		{
			name:        "Successful test #1",
			tokenString: validToken,
			secret:      validSecret,
			wantError:   false,
		},
		{
			name:        "Empty token #2",
			tokenString: "",
			secret:      validSecret,
			wantError:   true,
			errorText:   "token contains an invalid number of segments",
		},
		{
			name:        "Invalid token #3",
			tokenString: "invalid",
			secret:      validSecret,
			wantError:   true,
			errorText:   "token contains an invalid number of segments",
		},
		{
			name:        "Wrong secret #4",
			tokenString: validToken,
			secret:      []byte("wrong-secret-key"),
			wantError:   true,
			errorText:   "signature is invalid",
		},
		{
			name:        "Empty secret #5",
			tokenString: validToken,
			secret:      []byte(""),
			wantError:   true,
			errorText:   "signature is invalid",
		},
		{
			name:        "Nil secret #6",
			tokenString: validToken,
			secret:      nil,
			wantError:   true,
			errorText:   "signature is invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := ParseJWT(tc.tokenString, tc.secret)

			if tc.wantError {
				require.Error(t, err, "expected error but got none")
				if tc.errorText != "" {
					assert.Contains(t, err.Error(), tc.errorText, "unexpected error text")
				}
				assert.Nil(t, claims, "claims should be nil when error occurs")
				return
			}

			require.NoError(t, err, "unexpected error")
			require.NotNil(t, claims, "claims should not be nil")
			assert.Equal(t, validUserID, claims.UserID, "user ID in claims doesn't match")
		})
	}
}

const uuid = "0789b8d9-cef8-4837-be99-ec36fbf5c536"

var secret = []byte("secret")

func BenchmarkBuildJWTString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		//nolint:errcheck
		BuildJWT(uuid, secret)
	}
}

func BenchmarkParseJWT(b *testing.B) {
	jwt, err := BuildJWT(uuid, secret)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		//nolint:errcheck
		ParseJWT(jwt, secret)
	}
}
