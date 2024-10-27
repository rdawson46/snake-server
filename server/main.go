package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

/*
IDEA:
 - have tcp server and game 
 - get connections
 - update connections after game tick
 - send updates to all connections

*/

func assert(msg string, assertions ...bool) {
    for _, n := range assertions {
        if !n {
            fmt.Println(msg)
            syscall.Kill(syscall.Getpid(), syscall.SIGINT)
        }
    }
}

type serverConfig struct {
    Length  int
    Width   int
    maxConns int
}

type Packet struct {
    Version string `json:"version"`
    Length  int    `json:"length"`
    Width   int    `json:"width"`
    Page    []byte `json:"page"`
}

func encode(p Packet) ([]byte, error) {
    return json.Marshal(p)
}

func decode(b []byte) (*Packet, error) {
    p := &Packet{}
    err := json.Unmarshal(b, p)
    return p, err
}

func makePacket(s *server, b []byte) Packet {
    return Packet{
        Version: "0.1",
        Length: s.config.Length,
        Width: s.config.Width,
        Page: b,
    }
}

type connHandler func([]byte) (int, error)

type Node struct {
    value connHandler
    next  *Node
}

type list struct {
    head *Node
    tail *Node
    m    sync.Mutex
}

func newNode(c connHandler) *Node {
    return &Node{
        value: c,
        next: nil,
    }
}

func newList() list {
    return list{
        m: sync.Mutex{},
        head: nil,
        tail: nil,
    }
}

func (l *list) append(c connHandler) {
    n := newNode(c)

    l.m.Lock()
    defer l.m.Unlock()

    if l.head == nil {
        assert("tail isn't nil", l.tail == nil)

        l.head = n
        l.tail = n

        return
    } 

    assert("tail is nil, head is not", l.tail != nil)
    l.tail.next = n
    l.tail = n
}

// rework with pointers, seems unsafe
func (l *list) remove(prev, curr *Node) {
    l.m.Lock()
    defer l.m.Unlock()

    if prev == nil && curr == l.head {
        l.head = curr.next

        return
    }

    assert("prev is nil in remove", prev != nil)
    assert("curr is nil in remove", curr != nil)

    if curr == l.tail {
        l.tail = prev
    }

    prev.next = curr.next
}

func (l *list) removeHead() {
    l.m.Lock()
    defer l.m.Unlock()

    assert("head is nil in removeHead", l.head != nil)

    if l.head == l.tail {
        l.head = nil
        l.tail = nil
    }

    l.head = l.head.next
}

func (l *list) Write(b []byte) {
    var prev *Node = nil
    curr := l.head

    for curr != nil {
        _, err := curr.value(b)

        if err != nil {
            if errors.Is(err, net.ErrClosed) {
                if prev == nil && curr == l.head {
                    l.removeHead()
                } else {
                    l.remove(prev, curr)
                }

                continue
            }
        }

        prev = curr
        curr = curr.next
    }
}

type server struct {
    s        net.Listener
    shutdown chan struct{}
    conns    chan net.Conn
    l        list
    wg       sync.WaitGroup
    t        time.Ticker
    config   serverConfig
}


func newServer() (*server, error) {
    s, err := net.Listen("tcp", "127.0.0.1:8000")

    if err != nil {
        return nil, fmt.Errorf("failed to listen on server: %s", err.Error())
    }

    return &server{
        s: s,
        shutdown: make(chan struct{}),
        conns: make(chan net.Conn),
        l: newList(),
        t: *time.NewTicker(time.Second),
        config: serverConfig{
            Length: 2040,
            Width: 2040,
            maxConns: 10,
        },
    }, nil
}

func (s *server) start() {
    s.wg.Add(2)
    go s.handleConnections()
    go s.listen()
}

func (s *server) stop() {
    close(s.shutdown)
    s.s.Close()

    done := make(chan struct{})
    go func() {
        s.wg.Wait()
        close(done)
    }()

    select {
    case <- done:
        return
    case <- time.After(time.Second):
        fmt.Println("Timed out")
        return
    }
}

func (s *server) handleConnections() {
    defer s.wg.Done()

    for {
        select {
        case <- s.shutdown:
            return
        case <- s.t.C:
            fmt.Println("Sending")
            s.l.Write([]byte("Testing"))
        case conn, ok := <-s.conns:
            if !ok {
                fmt.Println("Error with conns chan")
                return
            }

            handler := s.handleConnection(conn)
            go s.l.append(handler)
        }
    }
}

func (s *server) handleConnection(c net.Conn) connHandler {
    return func(b []byte) (int, error) {
        return c.Write(b)
    }
}

// TODO: stalling issue comes here from waiting for next connection
func (s *server) listen() {
    defer s.wg.Done()

    fmt.Println("Listening...")

    for {
        select {
        case <- s.shutdown:
            return
        default:
            conn, err := s.s.Accept()

            if err != nil {
                if errors.Is(err, net.ErrClosed) {
                    fmt.Println("Connection closed")
                    continue
                }

                fmt.Printf("Error with connection: %s\n", err.Error())
                continue
            }

            s.conns <- conn
        }
    }
}

func run() {
    s, err := newServer()

    if err != nil {
        fmt.Printf("Error: %s", err.Error())
        os.Exit(1)
    }

    s.start()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <- sigChan

    fmt.Println("shutting down...")
    s.stop()
    fmt.Println("stopped")
}

func main() {
    fmt.Println("starting up...")
    run()
}
