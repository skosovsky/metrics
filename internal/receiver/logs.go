package receiver

import (
	"fmt"
	"net/http"
	"time"

	"metrics/internal/log"
)

func WithLogging(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var respData = new(responseData)

		logWriter := logResponseWriter{
			ResponseWriter: w,
			responseData:   respData,
		}

		next.ServeHTTP(&logWriter, r)

		duration := time.Since(start)

		log.Info("requestBody", //nolint:contextcheck // no ctx
			log.StringAttr("uri", r.RequestURI),
			log.StringAttr("method", r.Method),
			log.StringAttr("duration", duration.String()))

		log.Info("answer", //nolint:contextcheck // no ctx
			log.IntAttr("status", respData.statusCode),
			log.IntAttr("size", respData.size))
	}

	return http.HandlerFunc(logFn)
}

type (
	responseData struct {
		statusCode int
		size       int
	}

	logResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (l *logResponseWriter) Write(data []byte) (int, error) {
	size, err := l.ResponseWriter.Write(data)
	if err != nil {
		return 0, fmt.Errorf("error writing response: %w", err)
	}

	l.responseData.size += size

	return size, nil
}

func (l *logResponseWriter) WriteHeader(statusCode int) {
	l.ResponseWriter.WriteHeader(statusCode)
	l.responseData.statusCode = statusCode
}
