package internalhttp

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/logger"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler, logger *logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		latency := fmt.Sprintf("%d", time.Since(start).Milliseconds())
		clientIP := r.RemoteAddr
		if colon := len(clientIP) - 1; colon > 0 && clientIP[colon] == ']' {
			clientIP = clientIP[:colon]
		} else if colon := len(clientIP); colon > 0 {
			for i := colon - 1; i >= 0; i-- {
				if clientIP[i] == ':' {
					clientIP = clientIP[:i]
					break
				}
			}
		}
		date := time.Now().UTC().Format("[02/Jan/2006:15:04:05 -0700]")

		userAgent := r.UserAgent()
		if userAgent == "" {
			userAgent = `""`
		} else {
			userAgent = fmt.Sprintf("%q", userAgent)
		}
		code := strconv.Itoa(lrw.statusCode)
		uri := r.URL.RequestURI()
		logger.Info(
			"Request",
			"client_ip", clientIP,
			"date", date,
			"method", r.Method,
			"request_uri", uri,
			"proto", r.Proto,
			"code", code,
			"latency", latency,
			"user_agent", userAgent,
		)
	})
}
