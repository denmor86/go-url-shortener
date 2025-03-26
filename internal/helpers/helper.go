package helpers

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"time"
)

func MakeShortUrl(urlValue string, size int) string {

	data := fmt.Sprintf("%s%d", urlValue, time.Now().UnixNano())

	hash := md5.Sum([]byte(data))

	return base64.URLEncoding.EncodeToString(hash[:])[:size]
}
