package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	defaultAddr := envOrDefault("LICENSE_BIND_ADDR", ":8080")
	defaultDB := envOrDefault("LICENSE_DB_PATH", "data/licenses.db")
	defaultToken := os.Getenv("LICENSE_API_TOKEN")
	webhookSecret := os.Getenv("PAYSTACK_WEBHOOK_SECRET")

	addr := flag.String("addr", defaultAddr, "HTTP bind address")
	dbPath := flag.String("db", defaultDB, "path to the SQLite database file")
	token := flag.String("token", defaultToken, "shared installer token")
	paystackSecret := flag.String("paystack-webhook-secret", webhookSecret, "Paystack webhook signing secret")
	flag.Parse()

	store, err := NewLicenseStore(*dbPath)
	if err != nil {
		log.Fatalf("failed to open license store: %v", err)
	}
	defer store.Close()

	handler := NewLicenseHandler(store, *token, *paystackSecret)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/paystack/webhook", handler.HandlePaystackWebhook)
	mux.HandleFunc("/api/v1/licenses/validate", handler.ValidateLicense)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              *addr,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown error: %v", err)
		}
	}()

	log.Printf("license server listening on %s", *addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &logResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lrw.status, duration)
	})
}

type logResponseWriter struct {
	http.ResponseWriter
	status int
}

func (l *logResponseWriter) WriteHeader(statusCode int) {
	l.status = statusCode
	l.ResponseWriter.WriteHeader(statusCode)
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
