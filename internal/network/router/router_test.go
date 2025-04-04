package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	return resp
}

func TestHandleRouter(t *testing.T) {
	config := config.DefaultConfig()
	memstorage := memory.NewMemStorage()
	memstorage.Add("https://practicum.yandex.ru/", "12345678")
	memstorage.Add("https://google.com", "iFBc_bhG")

	ts := httptest.NewServer(HandleRouter(*config, memstorage))
	defer ts.Close()

	var testTable = []struct {
		url    string
		metod  string
		body   io.Reader
		status int
	}{
		// good
		{"/12345678", "GET", nil, http.StatusOK},
		{"/iFBc_bhG", "GET", nil, http.StatusOK},
		{"/", "POST", strings.NewReader("https://practicum.yandex.ru/"), http.StatusCreated},
		{"/", "POST", strings.NewReader("https://google.com"), http.StatusCreated},
		// bad
		{"/asdasdasd", "GET", nil, http.StatusBadRequest},
		{"/", "GET", nil, http.StatusMethodNotAllowed},
		{"/1234", "POST", strings.NewReader("https://practicum.yandex.ru/"), http.StatusMethodNotAllowed},
		{"/12345678/1234", "POST", strings.NewReader("https://practicum.yandex.ru/"), http.StatusNotFound},
	}
	for _, v := range testTable {
		resp := testRequest(t, ts, v.metod, v.url, v.body)
		assert.Equal(t, v.status, resp.StatusCode)
		defer resp.Body.Close()
	}
}
