package main

import "net"
import "fmt"
import "bufio"
import (
    "os"
    "gopkg.in/alecthomas/kingpin.v1"
)

var (
    addr = kingpin.Arg("addr", "Addres of node to connect to").Required().String()
)


func main() {
    kingpin.Parse();

    conn, err := net.Dial("tcp", *addr)

    if err != nil {
        fmt.Println("Couldn't connect to node")
        os.Exit(255)
    }

    fmt.Printf("Connected to node on %s\n", *addr)

    for {
        reader := bufio.NewReader(os.Stdin)
        fmt.Print("> ")

        text, _ := reader.ReadString('\n')
        fmt.Fprintf(conn, text)

        message, _ := bufio.NewReader(conn).ReadString('\n')
        fmt.Print(message)
    }
}