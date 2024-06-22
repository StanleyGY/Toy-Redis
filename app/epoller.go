package main

import (
	"log"
	"net"
	"reflect"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

type epollInfo struct {
	Fd    int
	Conns map[int]net.Conn
	Latch sync.Mutex
}

// One event loop per thread
type Epoller struct {
	epoll *epollInfo
}

func getFd(conn net.Conn) int {
	// Hacky way to get file descriptor of original conn
	fdVal := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("fd")
	fd := int(reflect.Indirect(fdVal).FieldByName("pfd").FieldByName("Sysfd").Int())
	return fd
}

func (el *Epoller) AddConn(conn net.Conn) error {
	fd := getFd(conn)

	// Events from original fd will now send to epoll fd
	err := unix.EpollCtl(
		el.epoll.Fd, syscall.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{
			Events: unix.POLLIN | unix.POLLHUP,
			Fd:     int32(fd),
		},
	)
	if err != nil {
		return err
	}

	// Map from fd to actual connection object
	el.epoll.Latch.Lock()
	defer el.epoll.Latch.Unlock()
	el.epoll.Conns[fd] = conn

	return nil
}

func (el *Epoller) Remove(conn net.Conn) error {
	defer conn.Close()

	// Tell OS that `epoll` should not look for this connection
	fd := getFd(conn)
	err := unix.EpollCtl(el.epoll.Fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}

	el.epoll.Latch.Lock()
	defer el.epoll.Latch.Unlock()
	delete(el.epoll.Conns, fd)

	return nil
}

func (el *Epoller) Wait() ([]net.Conn, error) {
	// Ask OS which connections (backed by fds) are ready for processing
	events := make([]unix.EpollEvent, 128)
	n, err := unix.EpollWait(el.epoll.Fd, events, -1)
	if err != nil {
		return nil, err
	}

	// Not sure if this can happen
	if n == 0 {
		log.Println("Received 0 events from epoll_wait")
		return nil, nil
	}

	el.epoll.Latch.Lock()
	defer el.epoll.Latch.Unlock()

	conns := make([]net.Conn, n)
	for i := 0; i < n; i++ {
		conns[i] = el.epoll.Conns[int(events[i].Fd)]
	}

	return conns, err
}

func MakeEpoller() (*Epoller, error) {
	// Tell OS to open an epoll queue
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	epollInfo := epollInfo{
		Conns: make(map[int]net.Conn),
		Fd:    fd,
		Latch: sync.Mutex{},
	}
	return &Epoller{epoll: &epollInfo}, nil
}
