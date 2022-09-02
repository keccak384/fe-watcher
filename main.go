package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

type server struct {
	redis Client
}

type Log struct {
	Contents string `json:"contents,omitempty"`
	Sources  string `json:"sources,omitempty"`
	Session  int64  `json:"session,omitempty"`
}

func (in *Log) Validate() error {
	if in == nil {
		return fmt.Errorf("log is nil")
	}
	if in.Contents == "" || in.Sources == "" || in.Session == 0 {
		return fmt.Errorf("log is empty")
	}
	return nil
}

func (s *server) LogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var l Log
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&l); err != nil {
		http.Error(w, fmt.Sprintf("read payload failed: %v", err), http.StatusInternalServerError)
		return
	}

	if err := l.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.redis.Set(fmt.Sprintf("log:%d", time.Now().Unix()), l, 0); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_, _ = fmt.Fprintf(w, "")
}

func main() {
	var err error
	s := &server{}
	if s.redis, err = NewClient(Options{Address: os.Getenv("REDIS_ADDRESS")}); err != nil {
		log.Fatalf("connect redis failed: %v", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/logs", s.LogHandler)

	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Printf("start listening at %s...\n", srv.Addr)
		if err = srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}
