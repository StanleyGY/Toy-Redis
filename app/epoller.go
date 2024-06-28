package main

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/unix"
)

type Epoller struct {
	ServerFd int
	EpollFd  int
	Conns    map[int]bool
}

func (el *Epoller) AddListener(serverFd int) error {
	// Register server fd to this epoll fd
	err := unix.EpollCtl(
		el.EpollFd, syscall.EPOLL_CTL_ADD, serverFd,
		&unix.EpollEvent{
			Events: unix.POLLIN | unix.POLLHUP,
			Fd:     int32(serverFd),
		},
	)
	if err != nil {
		return err
	}
	el.ServerFd = serverFd
	return nil
}

func (el *Epoller) AddConn(connfd int) error {
	_, found := el.Conns[connfd]
	if found {
		return fmt.Errorf("adding duplicate conn: %d", connfd)
	}
	el.Conns[connfd] = true

	// Events from original fd will now send to epoll fd
	return unix.EpollCtl(
		el.EpollFd, syscall.EPOLL_CTL_ADD, connfd,
		&unix.EpollEvent{
			Events: unix.POLLIN | unix.POLLHUP,
			Fd:     int32(connfd),
		},
	)
}

func (el *Epoller) RemoveConn(connfd int) error {
	_, found := el.Conns[connfd]
	if !found {
		return fmt.Errorf("removing non-existing conn: %d", connfd)
	}

	// `epoll` should not look for further events from this connection fd
	err := unix.EpollCtl(el.EpollFd, syscall.EPOLL_CTL_DEL, connfd, nil)
	if err != nil {
		return err
	}
	err = unix.Close(connfd)
	if err != nil {
		return err
	}
	delete(el.Conns, connfd)
	return nil
}

func (el *Epoller) GetEvents(timeout int) ([]unix.EpollEvent, error) {
	// Wait until some events are ready for processing
	var numEvents int
	var err error

	maxNumEvents := 128
	events := make([]unix.EpollEvent, maxNumEvents)
	for {
		numEvents, err = unix.EpollWait(el.EpollFd, events, timeout)
		// see remediation on EINTR error
		// https://stackoverflow.com/questions/6870158/epoll-wait-fails-due-to-eintr-how-to-remedy-this
		if err != unix.EINTR {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	if numEvents < maxNumEvents {
		events = events[:numEvents]
	}
	return events, nil
}

func MakeEpoller() (*Epoller, error) {
	epollFd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &Epoller{
		EpollFd: epollFd,
		Conns:   make(map[int]bool),
	}, nil
}
