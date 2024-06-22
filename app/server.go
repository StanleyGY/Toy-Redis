package main

import (
	"io"
	"log"
	"net"
)

func processIncomingRequest(epoller *Epoller, conn net.Conn) {
	buf := make([]byte, 1024)

	// Read data from connetion
	_, err := conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			err := epoller.Remove(conn)
			if err != nil {
				log.Println("Error closing connection: ", err.Error())
			} else {
				log.Println("Good bye!")
			}
			return
		}
		log.Println("Error reading from connection: ", err.Error())
	}

	// Process event
	output := []byte("+PONG\r\n")
	conn.Write(output)
}

func startServer() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		panic(err)
	}

	epoller, err := MakeEpoller()
	if err != nil {
		panic(err)
	}

	// Start a separate go routine to listen for events that are ready for processing
	go func() {
		for {
			conns, err := epoller.Wait()

			if err != nil {
				log.Println("Error epoller waiting for events: ", err.Error())
				continue
			}
			for _, conn := range conns {
				processIncomingRequest(epoller, conn)
			}
		}
	}()

	// Infinite loop for accepting connetions and adding to epoll queue
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			continue
		}

		err = epoller.AddConn(conn)
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			continue
		}
	}
}

func main() {
	startServer()
}
