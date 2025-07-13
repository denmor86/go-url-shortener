package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/denmor86/go-url-shortener.git/internal/logger"
)

// CompressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type CompressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// writerPool внутренний пул Writer
var writerPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(io.Discard)
	},
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	// Получаем Writer из пула (или создаём новый, если пул пуст)
	zw := writerPool.Get().(*gzip.Writer)

	// Сбрасываем Writer для использования с новым io.Writer
	zw.Reset(w)

	return &CompressWriter{
		w:  w,
		zw: zw,
	}
}

func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	contentType := c.w.Header().Get("Content-Type")
	switch {
	case strings.Contains(contentType, "application/json"),
		strings.Contains(contentType, "text/html"),
		strings.Contains(contentType, "text/plain"):
		return c.zw.Write(p)
	default:
		return c.w.Write(p)
	}
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < http.StatusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.

func (c *CompressWriter) Close() error {
	err := c.zw.Close()

	// Возвращаем Writer в пул для повторного использования
	writerPool.Put(c.zw)

	return err
}

// CompressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c CompressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// GzipHandle — middleware-gzip для HTTP-запросов.
func GzipHandle(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ow := w

		if supportsGzip(r.Header) {
			cw := NewCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		if sendsGzip(r.Header) {
			cr, err := NewCompressReader(r.Body)
			if err != nil {
				logger.Warn("Failed to create gzip reader:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	})
}

func supportsGzip(header http.Header) bool {
	return strings.Contains(header.Get("Accept-Encoding"), "gzip")
}

func sendsGzip(header http.Header) bool {
	return strings.Contains(header.Get("Content-Encoding"), "gzip")
}
