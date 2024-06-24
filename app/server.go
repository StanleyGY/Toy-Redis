package main

import (
	"io"
	"log"
	"net"

	"github.com/stanleygy/toy-redis/app/cmdexec"
	"github.com/stanleygy/toy-redis/app/parser"
	"github.com/stanleygy/toy-redis/app/resp"
)

func processIncomingRequest(epoller *Epoller, conn net.Conn) {
	buf := make([]byte, 1024)

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

	var outResp *resp.RespValue

	inResp, err := parser.Parse(buf)
	if err != nil {
		outResp = resp.MakeErorr(err.Error())
		log.Println("Error parsing Resp: ", err.Error())
		goto netwrite
	}

	outResp, err = cmdexec.Execute(inResp)
	if err != nil {
		outResp = resp.MakeErorr(err.Error())
		log.Println("Error executing Resp: ", err.Error())
		goto netwrite
	}

netwrite:
	conn.Write(outResp.ToByteArray())
}

func startServer() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		panic(err)
	}

	cmdexec.InitRedisDb()

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
