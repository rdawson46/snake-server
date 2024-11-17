package server

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"syscall"
	"time"
    "github.com/rdawson46/snake-server/packet"
)

// TODO: impl packet into the server/writing

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

type ServerConfig struct {
    Length  int
    Width   int
    maxConns int
}

type connHandler func([]byte) (int, error)

type Node struct {
    writer connHandler
    next  *Node
}

type list struct {
    head *Node
    tail *Node
    m    sync.Mutex
}

func newNode(c connHandler) *Node {
    return &Node{
        writer: c,
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
    l.m.Lock()
    defer l.m.Unlock()

    var prev *Node = nil
    curr := l.head

    for curr != nil {
        _, err := curr.writer(b)

        if err != nil {
            if errors.Is(err, net.ErrClosed) {
                if prev == nil && curr == l.head {
                    l.removeHead()
                } else {
                    l.remove(prev, curr)
                }
            } else {
                fmt.Println("Error occurred:", err.Error())
                continue
            }
        }

        prev = curr
        curr = curr.next
    }
}

type Server struct {
    s        net.Listener
    shutdown chan struct{}
    conns    chan net.Conn
    l        list
    wg       sync.WaitGroup
    t        time.Ticker
    Config   ServerConfig
}


func NewServer() (*Server, error) {
    s, err := net.Listen("tcp", "127.0.0.1:8000")

    if err != nil {
        return nil, fmt.Errorf("failed to listen on server: %s", err.Error())
    }

    return &Server{
        s: s,
        shutdown: make(chan struct{}),
        conns: make(chan net.Conn),
        l: newList(),
        t: *time.NewTicker(time.Second),
        Config: ServerConfig{
            Length: 2040,
            Width: 2040,
            maxConns: 10,
        },
    }, nil
}

func (s *Server) Start() {
    s.wg.Add(2)
    go s.handleConnections()
    go s.listen()
}

func (s *Server) Stop() {
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

func (s *Server) Write(b []byte) (int, error) {
    // make packet 
    // maybe return marshalled packet and send off bytes from here to list
    full, err := s.makePacket(string(b))

    if err != nil {
        fmt.Println("Error occured when writing:", err.Error())
        return 0, err
    }

    // write to list
    s.l.Write(full)
    return len(full), nil
}

func (s *Server) handleConnections() {
    defer s.wg.Done()

    for {
        select {
        case <- s.shutdown:
            return
        case <- s.t.C:
            fmt.Println("Sending")
            // TODO: use packets and send them here won't work here

            /*
            - remove tick and allow the game to write to this server
                - create a Write function for the server
            - make List public and allow for game to write straight to list
            */


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

func (s *Server) handleConnection(c net.Conn) connHandler {
    return func(b []byte) (int, error) {
        return c.Write(b)
    }
}

// TODO: stalling issue comes here from waiting for next connection
func (s *Server) listen() {
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

func (s *Server) makePacket(b string) ([]byte, error) {
    p := packet.Packet{
        Version: "0.1",
        Length: s.Config.Length,
        Width: s.Config.Width,
        Page: b,
    }

    return packet.Encode(p)
}

