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
func (s *server) handleConnection(c net.Conn) {}

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
