package main

import (
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

type connHandler func(string) (int, error)

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
        assert("tail isn't nil", l.tail != nil)

        l.head = n
        l.tail = n

        return
    } 

    assert("tail is nil, head is not", l.tail != nil)
    l.tail.next = n
    l.tail = n
}

// rework with pointers, remove extra looping
func (l *list) remove(i int) {
    l.m.Lock()
    defer l.m.Unlock()

    if l.head == nil {
        return
    }

    if i == 0 {
        if l.head == l.tail {
            l.tail = nil
        }

        l.head = l.head.next
        return
    }

    var prev *Node = nil
    curr := l.head
    currIndex := 0

    for currIndex != i && curr != nil {
        prev = curr
        curr = curr.next
    }

    if curr != nil {
        return
    }

    assert("prev is nil", prev != nil)
    prev.next = curr.next
}

func (l *list) cycle(s string) {
    // loop through list and i.connHandler(s)
    // if error remove at spot
}

type server struct {
    s        net.Listener
    shutdown chan struct{}
    conns    chan net.Conn
    wg       sync.WaitGroup
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
        case conn, ok := <-s.conns:
            if !ok {
                fmt.Println("Error with conns chan")
                return
            }

            // FIX: remove the go thread and adjust
            go s.handleConnection(conn)
        }
    }
}

/*
    TODO:
     - temp function
     - will have to figure out how to have all active connections giving and recving updates
     
     Things to consider 
      - how to store all connections
      - how to update all efficiently
      - how to get updates

    IDEA:
     - don't recv updates
     - check if connection is still alive
         - if so, send out game status
*/
func (s *server) handleConnection(c net.Conn) connHandler {
    return func(s string) (int, error) {
        return c.Write([]byte(s))
    }
}

func (s *server) listen() {
    defer s.wg.Done()

    for {
        select {
        case <- s.shutdown:
            return
        default:
            conn, err := s.s.Accept()

            if err != nil {
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
    run()
}
