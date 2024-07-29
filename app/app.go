package app

import (
	"fmt"
	"math/rand/v2"
	"time"
    "os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ScreenSaver struct {
    filled  [][]rune
    snakes  []snake
}

type direction int

const (
    left direction = iota
    up
    right 
    down
)

var Lines = [2]rune{
    '-',
    '|',
}

type snake struct {
    cur_x   int
    cur_y   int
    d     direction
    color lipgloss.Color // TODO: swap with a renderer
}

func newSnake() snake {
    return snake{
        cur_x: 0,
        cur_y: 0,
        d: right,
    }
}

func (s *snake) nextDirection() {
    odds := rand.IntN(101)

    // change direction
    if odds > 90 {
        // rotate left 90
        if s.d == left {
            s.d = down
        } else {
            s.d = s.d - 1
        }
    } else if odds > 80 {
        // rotate right 90
        if s.d == down {
            s.d = left
        } else {
            s.d = s.d + 1
        }
    }
}

func (s *snake) makeMove(g ScreenSaver) {
    switch s.d {
    case left:
        if s.cur_x == 0 {
            s.cur_x = len(g.filled[s.cur_y]) - 1
        } else {
            s.cur_x -= 1
        }
    case up:
        if s.cur_y == 0 {
            s.cur_y = len(g.filled) - 1
        } else {
            s.cur_y -= 1
        }
    case right:
        if s.cur_x == len(g.filled[s.cur_y]) - 1 {
            s.cur_x = 0
        } else {
            s.cur_x += 1
        }
    case down:
        if s.cur_y == len(g.filled) - 1 {
            s.cur_y = 0
        } else {
            s.cur_y += 1
        }
    }
}

func NewScreenSaver(snakes int) tea.Model {
    f := make([][]rune, 0)

    s := make([]snake, snakes)

    for i := range s {
        s[i] = newSnake()
    }

    return ScreenSaver{
        filled: f,
        snakes: s,
    }
}

type timer struct{}

func returnTimer(t time.Time) tea.Msg {
    return timer{}
}

func getNewDirection(s snake) snake {
    return s
}

func getRune(s snake) rune {
    switch s.d {
    case down:
        return 'v'
    case up:
        return '^'
    case right:
        return '>'
    case left:
        return '<'
    default:
        fmt.Println("Got impossible direction")
        os.Exit(1)
    }

    return 0
}

func (s ScreenSaver) Init() tea.Cmd {
    return tea.Tick(time.Millisecond * 250, returnTimer)
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
        for i, snake := range s.snakes {
            // get new direction
            snake.nextDirection()

            // move in direction/get next cell
            snake.makeMove(s)

            // apply rune
            s.filled[snake.cur_y][snake.cur_x] = getRune(snake)

            // TODO: add color

            s.snakes[i] = snake
        }

        return s, tea.Tick(time.Millisecond * 250, returnTimer)

    default:
        return s, nil
    }
}

func (s ScreenSaver) View() string { 
    // TODO: use render from snake
    render := lipgloss.NewStyle().Foreground(lipgloss.Color("#c565c5"))
    str := ""

    for _, v := range s.filled {
        str = fmt.Sprintf("%s\n%s", str, string(v))
    }

    return render.Render(str)
}
