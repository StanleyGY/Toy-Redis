package main

import (
	"log"

	"github.com/stanleygy/toy-redis/app/cmdexec"
	"github.com/stanleygy/toy-redis/app/resp"
	"golang.org/x/sys/unix"
)

func initListeners() *Epoller {
	epoller, err := MakeEpoller()
	if err != nil {
		panic(err)
	}

	serverFd, err := createServerFd()
	if err != nil {
		panic(err)
	}

	err = epoller.AddListener(serverFd)
	if err != nil {
		panic(err)
	}
	return epoller
}

func createServerFd() (int, error) {
	// Create a non-blocking fd for requests
	serverFd, err := unix.Socket(unix.AF_INET, unix.O_NONBLOCK|unix.SOCK_STREAM, 0)
	if err != nil {
		return 0, err
	}
	err = unix.SetNonblock(serverFd, true)
	if err != nil {
		return 0, err
	}

	// Bind server fd to addr and port
	serverAddr := &unix.SockaddrInet4{
		Port: 6379,
		Addr: [4]byte{0, 0, 0, 0},
	}
	err = unix.Bind(serverFd, serverAddr)
	if err != nil {
		return 0, err
	}
	err = unix.Listen(serverFd, 1024)
	if err != nil {
		return 0, err
	}
	return serverFd, nil
}

func processConnAcceptRequest(epoller *Epoller) {
	connfd, _, err := unix.Accept(epoller.ServerFd)
	if err != nil {
		log.Println(err.Error())
		return
	}
	epoller.AddConn(connfd)
}

func processConnReadRequest(connfd int, epoller *Epoller) {
	buf := make([]byte, 1024)

	numRead, err := unix.Read(connfd, buf)
	if err != nil {
		log.Println("Error reading from connection: ", err.Error())
		return
	}
	if numRead == 0 {
		// Connection closed for this socket
		err := epoller.RemoveConn(connfd)
		if err != nil {
			log.Println("Error closing connection: ", err.Error())
		} else {
			log.Println("Good bye!")
		}
		return
	}

	clientRequest := resp.Parse(buf)
	c := &cmdexec.ClientInfo{
		ConnFd:        connfd,
		ClientRequest: clientRequest,
	}
	cmdexec.Execute(c, clientRequest)
}

func processPostCmdExecutionEvents() {
	for _, ev := range cmdexec.EventBus {
		switch ev.Type {
		case cmdexec.EventReplyToClient:
			_, err := unix.Write(ev.Client.ConnFd, ev.Resp.ToByteArray())
			if err != nil {
				log.Println("Error writing Resp: ", err.Error())
			}
		case cmdexec.EventKeySpaceNotify:
			// Notify clients waiting on key space events
			// Queue unblocked clients for reprocessing
		}
	}
	cmdexec.Reset()
}

func reprocessUnblockedClients() {

}

func startServer() {
	epoller := initListeners()
	cmdexec.InitRedisDb()

	for {
		cmdexec.HandleBlockedClientsTimeout()

		// Listen for connection establishing events and other requests
		// TODO: caculate the timeout
		events, err := epoller.GetEvents(-1)
		if err != nil {
			log.Println("Error epoller waiting for events: ", err.Error())
			continue
		}
		for _, ev := range events {
			if ev.Fd == int32(epoller.ServerFd) {
				processConnAcceptRequest(epoller)
			} else {
				processConnReadRequest(int(ev.Fd), epoller)
			}
		}

		processPostCmdExecutionEvents()
		reprocessUnblockedClients()
	}
}

func main() {
	startServer()
}
