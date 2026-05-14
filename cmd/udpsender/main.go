package main

/*
	go run ./cmd/udpsender | tee /tmp/udpsender.txt
	nc -u -l 42069
*/

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"log"
)

func main() {
	conn, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		panic(err)
	}

	socket, err := net.DialUDP("udp", nil, conn)
	if err != nil {
		panic(err)
	}
	defer socket.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("> ")
		s, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error:", err)
			continue
		}
		n, err := socket.Write([]byte(s))
		if err != nil {
			log.Println("Error:", err)
		} else {
			log.Println("Sent", n, "bytes")
		}
	}
}
