package main

import (
	"fmt"
	"os"
    "github.com/rdawson46/snake-server/client/ui"

	tea "github.com/charmbracelet/bubbletea"
)

/*

IDEA:
 - connect to server
 - recv packet and decode
 - find current screen position and mask over packet.Page
 - display intersection

*/

func main() {
    p := tea.NewProgram(ui.NewUi())

    if _, err := p.Run(); err != nil {
        fmt.Println("Error occurred: ", err.Error())
        os.Exit(1)
    }
}
