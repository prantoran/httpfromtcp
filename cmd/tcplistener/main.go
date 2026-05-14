package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)

		str := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				if err == io.EOF {
					fmt.Println("-- End of file --")
					break
				} else {
					log.Fatal("Error reading file:", err)
				}
			}

			fmt.Printf("Read %d bytes: %s\n", n, string(data[:n]))

			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				data = data[i+1:]
				// fmt.Println(str)
				out <- str
				str = ""
			}

			str += string(data)

		}

		if len(str) > 0 {
			// fmt.Printf("Read: %s\n", str)
			out <- str
		}

	}()

	return out
}

func readFile(filename string) {
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal("Error opening file:", err)
	}
	defer f.Close()

	lines := getLinesChannel(f)
	for line := range lines {
		fmt.Printf("Read: %s\n", line)
	}
}

func readTCP() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("Error starting TCP server:", err)
	}
	defer listener.Close()
	fmt.Println("TCP server listening on port 42069")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error accepting connection:", err)
		}
		for line := range getLinesChannel(conn) {
			fmt.Printf("Read: %s\n", line)
		}
	}
}

func main() {
	// readFile("messages.txt")
	readTCP()
}
