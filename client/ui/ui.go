package ui

import (
	"fmt"
	"net"
	"os"

	tea "github.com/charmbracelet/bubbletea"
    "github.com/rdawson46/snake-server/packet"
)

// TODO: move packet struct to own package

type Ui struct {
    conn net.Conn
    p chan packet.Packet
    d chan deadMsg
}

func NewUi() Ui {
    conn, err := net.Dial("tcp", "127.0.0.1:8000")

    if err != nil {
        fmt.Println("Error when connection: ", err.Error())
        os.Exit(1)
    }

    return Ui{
        conn: conn,
    }
}

// MESSAGES
type serverMsg struct {
    packet packet.Packet
}

type deadMsg struct{}


// COMMANDS
func (u Ui) listen() tea.Cmd {
    return func() tea.Msg {
        return serverMsg {
            packet: <- u.p,
        }
    }
}

func (u Ui) dead() tea.Cmd {
    return func() tea.Msg {
        return <- u.d
    }
}

// EXTRA FUNCS
func (u Ui) listenOnConn() {
    // TODO: figure out size
    b := make([]byte, 256)
    n, err := u.conn.Read(b)
    // TODO: resume from here
}


// MODEL FUNCS
func (u Ui) Init() tea.Cmd {
    return tea.Batch(u.listen(), u.dead())
}

func (t Ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {return t, nil}

func (t Ui) View() string {return ""}
