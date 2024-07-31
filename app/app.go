package app

import (
	"fmt"
	"math/rand/v2"
	"time"
    "strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ScreenSaver struct {
    filled   [][]string
    snakes   []snake
    load     int
    capacity int
}

func (s ScreenSaver) full() bool {
    for _, v := range s.filled {
        for _, x := range v {
            if x == " " {
                return false
            }
        }
    }

    return true
}

type direction int

const (
    up direction = iota
    right 
    down
    left 
)

type snake struct {
    cur_x   int
    cur_y   int
    d     direction
    style lipgloss.Style
}

func newSnake(color lipgloss.TerminalColor) snake {
    return snake{
        cur_x: 0,
        cur_y: 0,
        d: right,
        // style: lipgloss.NewStyle().Background(color),
        style: lipgloss.NewStyle().Foreground(color),
    }
}

func (s *snake) nextDirection() {
    odds := rand.IntN(101)

    // change direction
    if odds >= 95 {
        // rotate left 90
        if s.d == up {
            s.d = left
        } else {
            s.d = s.d - 1
        }
    } else if odds >= 90 {
        // rotate right 90
        if s.d == left {
            s.d = up
        } else {
            s.d = s.d + 1
        }
    }
}

func (s *snake) changeColor() {
    p := rand.IntN(100)

    if p >= 90 {
        color := rand.IntN(0xffffff - 0xcccccc) + 0xcccccc

        s.style = lipgloss.NewStyle().Foreground(
            lipgloss.Color(fmt.Sprintf("#%d", color)),
        )
    }
}

func (s *snake) makeMove(g ScreenSaver) {
    switch s.d {
    case left:
        if s.cur_x == 0 {
            s.changeColor()
            s.cur_x = len(g.filled[s.cur_y]) - 1
        } else {
            s.cur_x -= 1
        }
    case up:
        if s.cur_y == 0 {
            s.changeColor()
            s.cur_y = len(g.filled) - 1
        } else {
            s.cur_y -= 1
        }
    case right:
        if s.cur_x == len(g.filled[s.cur_y]) - 1 {
            s.changeColor()
            s.cur_x = 0
        } else {
            s.cur_x += 1
        }
    case down:
        if s.cur_y == len(g.filled) - 1 {
            s.changeColor()
            s.cur_y = 0
        } else {
            s.cur_y += 1
        }
    }
}

func NewScreenSaver(snakes int) tea.Model {
    f := make([][]string, 0)

    s := make([]snake, snakes)

    for i := range s {
        color := rand.IntN(0xffffff - 0xcccccc) + 0xcccccc
        s[i] = newSnake(lipgloss.Color(fmt.Sprintf("#%d", color)))
    }

    return ScreenSaver{
        filled: f,
        snakes: s,
        capacity: 2500,
        load: 0,
    }
}

func getNewDirection(s snake) snake {
    return s
}

var Lines = [16]string{
    "┃",
    "┏",
    " ",
    "┓",
    "┛",
    "━",
    "┓",
    " ",
    " ",
    "┗",
    "┃",
    "┛",
    "┗",
    " ",
    "┏",
    "━",
}

func getString(s snake, prev direction) string {
    return Lines[ s.d + prev * 4 ]
}

func (s ScreenSaver) Init() tea.Cmd {
    return tea.Tick(time.Millisecond * 5, returnTimer)
}

type timer struct{}

func returnTimer(t time.Time) tea.Msg {
    return timer{}
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
        f := make([][]string, msg.Height)

        for i := range f {
            f[i] = make([]string, msg.Width)
            for j := range f[i] {
                f[i][j] = " "
            }
        }

        s.filled = f
        return s, nil

    case timer:
        for i, snake := range s.snakes {
            temp_dir := snake.d
            snake.nextDirection()

            if snake.d != temp_dir {
                s.filled[snake.cur_y][snake.cur_x] = snake.style.Render(getString(snake, temp_dir))
                snake.makeMove(s)
            }

            s.filled[snake.cur_y][snake.cur_x] = snake.style.Render(getString(snake, snake.d))
            snake.makeMove(s)

            s.snakes[i] = snake
            s.load += 1
        }

        if s.full() || s.load == s.capacity {
            y := make([][]string, len(s.filled))

            for x := range y {
                y[x] = make([]string, len(s.filled[0]))

                for i := range y[x] {
                    y[x][i] = " "
                }
            }

            s.filled = y
            s.load = 0
        }

        return s, tea.Tick(time.Millisecond * 5, returnTimer)

    default:
        return s, nil
    }
}

func (s ScreenSaver) View() string { 
    str := ""

    for _, v := range s.filled {
        temp := strings.Join(v, "")
        str = fmt.Sprintf("%s\n%s", str, temp)
    }

    return str
}
