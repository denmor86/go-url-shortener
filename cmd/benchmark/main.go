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

	"github.com/denmor86/go-url-shortener/internal/usecase"
	"github.com/denmor86/go-url-shortener/pkg/random"
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

		// создаем клиента
		client := makeClient()

		// базовый адрес
		baseURL := makeURL(address, port)

		// проверяем коннект
		checkPing(client, baseURL)

		// проверяем создание сокращенных URL по полному
		checkPostURL(client, baseURL)

		// генерируем массив URL и отправляем его для создания
		urls := checkSendBatch(client, baseURL)

		// проверяем запросы ранее созданных записей
		deleteItems := checkGetURL(client, urls)

		// проверяем запрос на получение всех записей пользователя
		checkGetURLS(client, baseURL)

		// проверяем запрос на удаление записей
		checkDeleteURLS(client, baseURL, deleteItems)

		// ожидание
		fmt.Printf("waiting... \n")
		time.Sleep(time.Second)
	}
}

// makeClient - формирование http клиента с пустым cookie
func makeClient() *http.Client {
	// создаем пустую cookie
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Error creating cookie jar: %v", err)
	}
	return &http.Client{
		Timeout: timeout,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Возвращаем эту ошибку, чтобы остановить редирект
			return http.ErrUseLastResponse
		},
	}
}

// makeURL - формирование базового адреса сервиса
func makeURL(address string, port uint16) string {
	return fmt.Sprintf("http://%s:%d", address, port)
}

// checkPing - метод для проверки GET метода ping
func checkPing(client *http.Client, baseURL string) {
	_, err := sendPing(client, baseURL)
	if err != nil {
		log.Fatalf("Error ping: %v", err)
	}
}

// checkPostURL - метод для проверки POST метода /
func checkPostURL(client *http.Client, baseURL string) {
	for i := 1; i <= batchSize; i++ {
		URL, sendError := sendPostURL(client, baseURL, random.URL().String())
		if sendError != nil {
			log.Fatalf("Send post URL failed: %v\n", sendError)
		}
		fmt.Printf("Base URL: %s\n", URL)
	}
}

// checkSendBatch - метод проверки метода api/shorten/batch (возвращает массис созданных записей)
func checkSendBatch(client *http.Client, baseURL string) []string {
	batch := GenerateURLItems(batchSize)
	batchResponse, sendError := sendBatch(client, baseURL, batch)
	if sendError != nil {
		log.Fatalf("Send batch failed: %v", sendError)
	}
	// разбираем ответ
	urls, err := parseBatchResponse(batchResponse)
	if err != nil {
		log.Fatalf("parse batch responce failed: %v", err)
	}
	return urls
}

// checkGetURL - метод проверки GET запроса коротких URL (возвращает массив записей для удаления)
func checkGetURL(client *http.Client, urls []string) []string {
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
	return deleteItems
}

// checkGetURLS - метод проверки GET запроса всех URL пользователя (по токену)
func checkGetURLS(client *http.Client, baseURL string) {
	resp, err := sendGetURLS(client, baseURL)
	if err != nil {
		log.Fatalf("Send get URLs failed: %v", err)
	}
	fmt.Printf("URLs: %s\n", resp)
}

// checkDeleteURLS - метод проверки запроса удаления записей пользователя (по токену)
func checkDeleteURLS(client *http.Client, baseURL string, deleteItems []string) {
	_, err := sendBachDelete(client, baseURL, deleteItems)
	if err != nil {
		log.Fatalf("Send batch delete failed: %v", err)
	}
}

// sendPing отправляет PING запрос
func sendPing(client *http.Client, baseURL string) (string, error) {
	url := fmt.Sprintf("%s/ping", baseURL)
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET request to %s error: %v", url, err)
	}
	defer resp.Body.Close()

	return handleResponse(resp)
}

// sendPostURL отправляет Post запрос c коротким адресом
func sendPostURL(client *http.Client, baseURL string, URL string) (string, error) {
	resp, err := client.Post(baseURL, "text/plain", bytes.NewBufferString(URL))
	if err != nil {
		return "", fmt.Errorf("GET request to %s error: %v", URL, err)
	}
	defer resp.Body.Close()

	return handleResponse(resp)
}

// sendBach отправляет POST запрос с JSON (Batch) телом
func sendBatch(client *http.Client, baseURL string, data []usecase.RequestItem) (string, error) {
	url := fmt.Sprintf("%s/api/shorten/batch", baseURL)
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
func sendBachDelete(client *http.Client, baseURL string, data []string) (string, error) {

	url := fmt.Sprintf("%s/api/user/urls", baseURL)
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
func sendGetURLS(client *http.Client, baseURL string) (string, error) {
	url := fmt.Sprintf("%s/api/user/urls", baseURL)

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
