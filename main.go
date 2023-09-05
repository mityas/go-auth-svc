package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const SERVER_PORT = "9000"

func testEndpoint(endpoint string) string {
	res, err := http.Get("http://localhost:" + SERVER_PORT + "/" + endpoint)
	if err != nil {
		log.Println(err)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}

	return string(b)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/signup", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "signup endpoint")
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "login endpoint")
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "token endpoint")
	})
	mux.HandleFunc("/refresh", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "refresh endpoint")
	})

	srv := &http.Server{
		Addr:           ":" + SERVER_PORT,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}

	// Graceful shutdown
	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Println("Shutdown...")
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Println("Starting server...")
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	// Test endpoints
	cases := []struct {
		in, want string
	}{
		{"signup", "signup endpoint"},
		{"login", "login endpoint"},
		{"token", "token endpoint"},
		{"refresh", "refresh endpoint"},
	}
	for _, c := range cases {
		got := testEndpoint(c.in)
		if got != c.want {
			log.Printf("testEndpoint(%q) == %q, want %q\n", c.in, got, c.want)
		} else {
			log.Printf("testEndpoint(%q) == %q\n", c.in, got)
		}
	}

	<-idleConnsClosed
}
