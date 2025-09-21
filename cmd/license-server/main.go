package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/umar/amazon-product-scraper/internal/licensing"
)

func main() {
	var (
		addr         = flag.String("addr", ":8080", "listen address for the HTTP server")
		dbPath       = flag.String("db", "licenses.db", "path to the SQLite database")
		expiryDays   = flag.Int("expiry", 365, "default license validity in days (0 for no expiry)")
		readTimeout  = flag.Duration("read-timeout", 10*time.Second, "HTTP server read timeout")
		writeTimeout = flag.Duration("write-timeout", 10*time.Second, "HTTP server write timeout")
		idleTimeout  = flag.Duration("idle-timeout", 60*time.Second, "HTTP server idle timeout")
	)
	flag.Parse()

	defaultExpiry := time.Duration(*expiryDays) * 24 * time.Hour
	service, err := licensing.NewService(*dbPath, defaultExpiry)
	if err != nil {
		log.Fatalf("failed to initialize licensing service: %v", err)
	}
	defer service.Close()

	api := licensing.NewAPI(service)
	mux := http.NewServeMux()
	api.Register(mux)

	server := &http.Server{
		Addr:         *addr,
		Handler:      withLogging(mux),
		ReadTimeout:  *readTimeout,
		WriteTimeout: *writeTimeout,
		IdleTimeout:  *idleTimeout,
	}

	go func() {
		log.Printf("license server listening on %s (db: %s)", *addr, *dbPath)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown on Ctrl+C / SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("shutting down license server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, lrw.status, duration)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.status = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.status == 0 {
		lrw.status = http.StatusOK
	}
	return lrw.ResponseWriter.Write(b)
}
