package helpers

import (
	"testing"

	"github.com/denmor86/go-url-shortener.git/pkg/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeShortURL(t *testing.T) {
	t.Run("Successful", func(t *testing.T) {
		testCases := []struct {
			name     string
			url      string
			size     int
			expected int
		}{
			{"Normal URL", "https://example.com", 8, 8},
			{"Minimum size", "https://example.com", 1, 1},
			{"Maximum size", "https://example.com", 24, 24},
			{"URL with special chars", "https://example.com/тест?param=value", 10, 10},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := MakeShortURL(tc.url, tc.size)
				require.NoError(t, err)
				assert.Len(t, result, tc.expected)
				assertValidBase64URL(t, result)
			})
		}
	})

	t.Run("Errors", func(t *testing.T) {
		testCases := []struct {
			name    string
			url     string
			size    int
			errText string
		}{
			{"Empty URL", "", 8, "url cannot be empty"},
			{"Zero size", "https://example.com", 0, "size must be positive, got 0"},
			{"Negative size", "https://example.com", -1, "size must be positive, got -1"},
			{"Size too large", "https://example.com", 25, "size too large, maximum is 24, got 25"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := MakeShortURL(tc.url, tc.size)
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errText)
				assert.Empty(t, result)
			})
		}
	})
}

func assertValidBase64URL(t *testing.T, s string) {
	t.Helper()
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_="
	for _, c := range s {
		assert.Contains(t, validChars, string(c), "URL contains invalid character")
	}
}

const size = 8

func BenchmarkMakeShortURL(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		//nolint:errcheck
		MakeShortURL(random.URL().String(), size)
	}
}
