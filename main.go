package main

import (
	"fmt"
	"os"
    "syscall"
    "os/signal"
    tea "github.com/charmbracelet/bubbletea"

	"github.com/rdawson46/snake-server/server"
	"github.com/rdawson46/snake-server/game"
)

/* 
TODO:
 - get server screen size
 - apply location to client
*/

func main() {
    s, err := server.NewServer()

    if err != nil {
        fmt.Printf("Error: %s", err.Error())
        os.Exit(1)
    }

    s.Start()

    go func() {
        program := tea.NewProgram(game.NewScreenSaver(2), tea.WithOutput(s))

        if _, err := program.Run(); err != nil {
            fmt.Println("Error with model:", err.Error())
            os.Kill.Signal()
        }
    }()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <- sigChan

    fmt.Println("shutting down...")
    s.Stop()
    fmt.Println("stopped")
}
