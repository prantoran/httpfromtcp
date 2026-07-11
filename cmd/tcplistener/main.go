/*
	Run the TCP listener and pipe the output to a file:
		go run ./cmd/tcplistener | tee /tmp/requestline.txt

	Test parsing:
		curl http://localhost:42069
		curl -X POST http://localhost:42069 -d '{"model": "a32", "type": "home"}'
*/

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/prantoran/httpfromtcp/internal/request"
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

			// fmt.Printf("Read %d bytes: %s\n", n, string(data[:n]))

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
		// for line := range getLinesChannel(conn) {
		// 	fmt.Printf("Read: %s\n", line)
		// }
		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("Error reading request:", err)
		}
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Printf("Headers:\n")
		r.Headers.ForEach(func(key, value string) {
			fmt.Printf("- %s: %s\n", key, value)
		})

		if request.GetIntHeader(r.Headers, "content-length", 0) > 0 {
			fmt.Printf("Body:\n")
			fmt.Printf("%s\n", r.Body)
		}

		conn.Close()
	}
}

func main() {
	// readFile("messages.txt")
	readTCP()
}
