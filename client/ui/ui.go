package ui

import (
	"fmt"
	"net"
	"os"
    "errors"

	tea "github.com/charmbracelet/bubbletea"
    "github.com/rdawson46/snake-server/packet"
)

// TODO: move packet struct to own package

type Ui struct {
    conn net.Conn
    p chan *packet.Packet
    d chan deadMsg
    page string
    serverWidth int
    serverHeight int
    width int
    height int
}

func NewUi() Ui {
    conn, err := net.Dial("tcp", "127.0.0.1:8000")

    d := make(chan deadMsg)
    p := make(chan *packet.Packet)

    if err != nil {
        fmt.Println("Error when connection: ", err.Error())
        os.Exit(1)
    }

    return Ui{
        conn: conn,
        page: "",
        p: p,
        d: d,
        serverWidth: 0,
        serverHeight: 0,
        width: 0,
        height: 0,
    }
}

// MESSAGES
type serverMsg struct {
    packet *packet.Packet
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
    b := make([]byte, 16248)

    for {
        n, err := u.conn.Read(b)

        if err != nil {
            if errors.Is(err, net.ErrClosed) {
                u.d <- deadMsg{}
            } else {
                continue
            }
        }

        p, err := packet.Decode(b[:n])

        if err != nil {
            continue
        }

        u.p <- p
    }
}


func handlePacket(ui Ui, p *packet.Packet) Ui {
    ui.serverWidth = p.Width
    ui.serverHeight = p.Length

    // will need to figure out where in the page should be shown
    ui.page = p.Page

    return ui
}

// MODEL FUNCS
func (u Ui) Init() tea.Cmd {
    return tea.Batch(u.listen(), u.dead())
}

func (u Ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        u.height = msg.Height
        u.width = msg.Width
        return u, nil
    case deadMsg:
        return u, tea.Quit
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return u, tea.Quit
        }
    case serverMsg:
        u = handlePacket(u, msg.packet)
        return u, u.listen()

    }

    return u, nil
}

func (u Ui) View() string {
    return u.page
}
