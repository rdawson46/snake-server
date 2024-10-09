package main

import (
    "net"
    "fmt"
    "time"
    "os"
    "sync"
)

func test_simple_client(id int, wg *sync.WaitGroup) {
    defer wg.Done()

    conn, err := net.Dial("tcp", "127.0.0.1:8000")

    if err != nil {
        fmt.Println("Hit error when trying to connect")
        fmt.Println(err.Error())
        os.Exit(1)
    }

    fmt.Printf("Listening with thread %d\n", id)

    for range 5 {
        b := make([]byte, 512)
        n, err := conn.Read(b)

        if err != nil {
            fmt.Println("Hit error when listening")
            fmt.Println(err.Error())
            break
        }

        fmt.Printf("%d recv: %s\n", id, b[:n])
    }
}

func main() {
    t := time.NewTicker(time.Second * 2)

    wg := &sync.WaitGroup{}

    for i := range 10 {
        go test_simple_client(i, wg)
        wg.Add(1)
        <- t.C
    }

    wg.Wait()
}
