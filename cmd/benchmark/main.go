package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"time"

	"github.com/denmor86/go-url-shortener.git/internal/usecase"
	"github.com/denmor86/go-url-shortener.git/pkg/random"
)

// GenerateURLItems создает срез RequestItem указанного размера
func GenerateURLItems(count int) []usecase.RequestItem {
	if count < 1 {
		return []usecase.RequestItem{}
	}

	items := make([]usecase.RequestItem, 0, count)
	for i := 1; i <= count; i++ {
		items = append(items, usecase.RequestItem{
			ID:  fmt.Sprintf("%d", i),
			URL: random.URL().String(),
		})
	}
	return items
}

const (
	address   = "localhost"
	port      = 8080
	timeout   = 10 * time.Second
	batchSize = 10000
	runCount  = 10
)

func main() {
	for i := 1; i <= runCount; i++ {
		// создаем пустую cookie
		jar, err := cookiejar.New(nil)
		if err != nil {
			log.Fatalf("Error creating cookie jar: %v", err)
		}
		// создаем клиента
		client := &http.Client{
			Timeout: timeout,
			Jar:     jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Возвращаем эту ошибку, чтобы остановить редирект
				return http.ErrUseLastResponse
			},
		}
		// проверяем коннект
		_, err = sendPing(client, address, port)
		if err != nil {
			log.Fatalf("Error ping: %v", err)
		}
		// запрашиваем создание сокращенный URL по каждому полному
		for i := 1; i <= batchSize; i++ {
			URL, err := sendPostURL(client, address, port, random.URL().String())
			if err != nil {
				log.Fatalf("Send post URL failed: %v\n", err)
			}
			fmt.Printf("Base URL: %s\n", URL)
		}

		// генерируем массив и отправляем его
		batch := GenerateURLItems(batchSize)
		batchResponse, err := sendBatch(client, address, port, batch)
		if err != nil {
			log.Fatalf("Send batch failed: %v", err)
		}
		// разбираем ответ
		urls, err := parseBatchResponse(batchResponse)
		deleteItems := make([]string, 0, len(urls))
		for _, url := range urls {
			// подготовка записей для удаления
			deleteItems = append(deleteItems, getShortSegment(url))

			// запрашиваем URL по каждому сокращенному
			URL, err := sendGetURL(client, url)
			if err != nil {
				log.Fatalf("Send get URL failed: %v\n", err)
			}
			fmt.Printf("Short URL: %s\n", URL)
		}

		// отправляем запрос на получение всех записей
		resp, err := sendGetURLS(client, address, port)
		if err != nil {
			log.Fatalf("Send get URLs failed: %v", err)
		}
		fmt.Printf("URLs: %s\n", resp)

		// кидаем запрос на удаление
		_, err = sendBachDelete(client, address, port, deleteItems)
		if err != nil {
			log.Fatalf("Send batch delete failed: %v", err)
		}
		fmt.Printf("waiting... \n")
		time.Sleep(time.Second)
	}
}

// sendPing отправляет PING запрос
func sendPing(client *http.Client, address string, port uint16) (string, error) {
	url := fmt.Sprintf("http://%s:%d/ping", address, port)
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET request to %s error: %v", url, err)
	}
	defer resp.Body.Close()

	return handleResponse(resp)
}

// sendBach отправляет POST запрос с JSON (Batch) телом
func sendBatch(client *http.Client, address string, port uint16, data []usecase.RequestItem) (string, error) {
	url := fmt.Sprintf("http://%s:%d/api/shorten/batch", address, port)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("POST request to %s error: %v", url, err)
	}
	defer resp.Body.Close()

	return handleResponse(resp)
}

// sendBachDelete отправляет DELETE запрос с JSON (Batch) телом
func sendBachDelete(client *http.Client, address string, port uint16, data []string) (string, error) {

	url := fmt.Sprintf("http://%s:%d/api/user/urls", address, port)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating DELETE request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("DELETE urls to %s error: %v", url, err)
	}
	defer resp.Body.Close()

	return handleResponse(resp)
}

// sendPostURL отправляет Post запрос c коротким адресом
func sendPostURL(client *http.Client, address string, port uint16, URL string) (string, error) {
	url := fmt.Sprintf("http://%s:%d/", address, port)
	resp, err := client.Post(url, "text/plain", bytes.NewBufferString(URL))
	if err != nil {
		return "", fmt.Errorf("GET request to %s error: %v", URL, err)
	}
	defer resp.Body.Close()

	return handleResponse(resp)
}

// sendGetURL отправляет GET запрос c коротким адресом
func sendGetURL(client *http.Client, URL string) (string, error) {
	resp, err := client.Get(URL)
	if err != nil {
		return "", fmt.Errorf("GET request to %s error: %v", URL, err)
	}
	defer resp.Body.Close()

	return handleResponse(resp)
}

// sendGetURLS отправляет GET запроc c получением всех записей пользователя
func sendGetURLS(client *http.Client, address string, port uint16) (string, error) {
	url := fmt.Sprintf("http://%s:%d/api/user/urls", address, port)

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET request to %s error: %v", url, err)
	}
	defer resp.Body.Close()

	return handleResponse(resp)
}

// handleResponse обрабатывает HTTP ответ
func handleResponse(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	switch resp.StatusCode {
	case http.StatusCreated:
	case http.StatusOK:
	case http.StatusTemporaryRedirect:
	case http.StatusAccepted:
	default:
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	return string(body), nil
}

// parseBatchResponse обрабатывает batch
func parseBatchResponse(resp string) ([]string, error) {

	var respItems []usecase.ResponseItem
	if err := json.Unmarshal([]byte(resp), &respItems); err != nil {
		return nil, fmt.Errorf("error unmarshal body: %w", err)
	}

	if len(respItems) == 0 {
		return nil, fmt.Errorf("empty batch response")
	}

	items := make([]string, 0, len(respItems))
	for _, item := range respItems {
		items = append(items, item.URL)
	}

	return items, nil
}

func getShortSegment(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return path.Base(u.Path)
}
