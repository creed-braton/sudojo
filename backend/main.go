package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"sudojo/server"
)

func main() {
	fmt.Println("Starting Sudoku game server...")

	s := server.NewServer()
	mux := http.NewServeMux()
	s.SetupRoutes(mux)

	go func() {
		addr := ":8080"
		fmt.Printf("Server running at http://localhost%s\n", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Server shutting down...")
}
