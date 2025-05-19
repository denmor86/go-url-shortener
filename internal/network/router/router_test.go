package router

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denmor86/go-url-shortener.git/internal/config"
	"github.com/denmor86/go-url-shortener.git/internal/logger"
	"github.com/denmor86/go-url-shortener.git/internal/storage"
	"github.com/denmor86/go-url-shortener.git/internal/usecase"
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
	if err := logger.Initialize(config.LogLevel); err != nil {
		logger.Panic(err)
	}
	defer logger.Sync()
	store := storage.NewStorage(config)
	store.AddRecord(context.Background(), storage.TableRecord{OriginalURL: "https://practicum.yandex.ru/", ShortURL: "12345678"})
	store.AddRecord(context.Background(), storage.TableRecord{OriginalURL: "https://google.com", ShortURL: "iFBc_bhG"})

	usecase := usecase.NewUsecase(config, store)

	ts := httptest.NewServer(HandleRouter(usecase))
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
		{"/api/shorten", "POST", strings.NewReader("{\"url\": \"https://practicum.yandex.ru\"}"), http.StatusCreated},
		{"/api/shorten", "POST", strings.NewReader("{\"url\": \"https://google.com\", \"test\": \"test message\"}"), http.StatusCreated},
		{"/api/shorten/batch", "POST", strings.NewReader(`[{"correlation_id":"c978edb7-eb81-45b3-bcc7-e5cf9f5781cd","original_url":"http://qpabthuzw1vjfl.com"},{"correlation_id":"0a5c6e26-f875-44c8-9e09-1ceefa82235e","original_url":"http://nqea9x1nxhuinc.biz/cvn6iupy"}]`), http.StatusCreated},
		{"/api/user/urls", "GET", nil, http.StatusNoContent},
		{"/ping", "GET", nil, http.StatusOK},
		// bad
		{"/asdasdasd", "GET", nil, http.StatusBadRequest},
		{"/", "GET", nil, http.StatusMethodNotAllowed},
		{"/1234", "POST", strings.NewReader("https://practicum.yandex.ru/"), http.StatusMethodNotAllowed},
		{"/12345678/1234", "POST", strings.NewReader("https://practicum.yandex.ru/"), http.StatusNotFound},
		{"/api/shorten", "POST", strings.NewReader("{\"test\": \"https://practicum.yandex.ru\"}"), http.StatusBadRequest},
		{"/api/shorten", "POST", strings.NewReader("<request><url>google.com</url></request>"), http.StatusBadRequest},
		{"/api/shorten1", "POST", strings.NewReader("{\"url\": \"https://practicum.yandex.ru\"}"), http.StatusNotFound},
		{"/api/shorten/batch", "POST", strings.NewReader(`[{"correlation_id":"c978edb7-eb81-45b3-bcc7-e5cf9f5781cd","original_url":""},{"correlation_id":"0a5c6e26-f875-44c8-9e09-1ceefa82235e","original_url":""}]`), http.StatusBadRequest},
	}
	for _, v := range testTable {
		resp := testRequest(t, ts, v.metod, v.url, v.body)
		assert.Equal(t, v.status, resp.StatusCode)
		defer resp.Body.Close()
	}
}
