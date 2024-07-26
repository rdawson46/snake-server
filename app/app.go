package app

import (
	"fmt"
	"math/rand/v2"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ScreenSaver struct {
    filled  [][]rune
}

type snake struct {
    body  []uint
    color lipgloss.Color
}

func NewScreenSaver() tea.Model {
    f := make([][]rune, 0)

    return ScreenSaver{
        filled: f,
    }
}

type timer struct{}

func returnTimer(t time.Time) tea.Msg {
    return timer{}
}

func (s ScreenSaver) Init() tea.Cmd {
    return tea.Tick(time.Second, returnTimer)
}

func (s ScreenSaver) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return s, tea.Quit
        default:
            return s, nil
        }

    case tea.WindowSizeMsg:
        f := make([][]rune, msg.Height)

        for i := range f {
            f[i] = make([]rune, msg.Width)
            for j := range f[i] {
                f[i][j] = ' '
            }
        }

        s.filled = f
        return s, nil

    case timer:
        r1 := rand.IntN(len(s.filled))
        r2 := rand.IntN(len(s.filled[r1]))
        s.filled[r1][r2] = 'X'
        return s, tea.Tick(time.Second, returnTimer)

    default:
        return s, nil
    }
}

func (s ScreenSaver) View() string {
    str := ""

    for _, v := range s.filled {
        str = fmt.Sprintf("%s\n%s", str, string(v))
    }

    return str
}
