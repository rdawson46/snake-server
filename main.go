package main

import (
	"fmt"
	"os"
    "syscall"
    "os/signal"

	"github.com/rdawson46/snake-server/server"
)

// main func for running server
func main() {
    s, err := server.NewServer()

    if err != nil {
        fmt.Printf("Error: %s", err.Error())
        os.Exit(1)
    }

    s.Start()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <- sigChan

    fmt.Println("shutting down...")
    s.Stop()
    fmt.Println("stopped")
}
