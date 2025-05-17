package helpers

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

func MakeShortURL(urlValue string, size int) string {

	data := fmt.Sprintf("%s%d", urlValue, time.Now().UnixNano())

	hash := md5.Sum([]byte(data))

	return base64.URLEncoding.EncodeToString(hash[:])[:size]
}

func MakeURL(baseURL, shortURL string) string {

	var fullURL string
	fullURL = baseURL
	if !strings.HasSuffix(fullURL, "/") {
		fullURL += "/"
	}
	fullURL += shortURL
	return fullURL
}
