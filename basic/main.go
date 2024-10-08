package main

import (
    "fmt"
    "os"
    "github.com/rdawson46/screensaver/app"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    program := tea.NewProgram(app.NewScreenSaver(5), tea.WithAltScreen())

    if _, err := program.Run(); err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }
}
