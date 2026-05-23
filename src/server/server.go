package server

import (
	"encoding/json"
	"fmt"
	"syscall"

	"github.com/orayew2002/db/src/cli"
	"github.com/orayew2002/db/src/db"
	"github.com/orayew2002/db/src/executor"
)

type Config struct {
	Port    int
	Addr    [4]byte
	Backlog int
}

type Server struct {
	fd   int
	exec *executor.Exec
}

func Default(db *db.Database) (*Server, error) {
	return New(Config{
		Port:    9696,
		Addr:    [4]byte{0, 0, 0, 0},
		Backlog: 10,
	}, db)
}

func New(s Config, db *db.Database) (*Server, error) {
	// 1. socket()
	// We ask the Linux kernel to create a TCP socket.
	//
	// A socket is NOT a network connection yet.
	// It is just a kernel object (endpoint) that can later handle TCP communication.
	//
	// The kernel returns a file descriptor (fd),
	// which is just an integer handle to this socket.
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, fmt.Errorf("error opening tcp socket: %w", err)
	}

	// 2. bind()
	// We tell the kernel:
	// "Attach this socket to a specific IP address and port."
	//
	// Addr: s.Addr means:
	// listen on ALL network interfaces (localhost, LAN, public IP, etc.)
	//
	// After bind():
	// any incoming TCP packet sent to port 9696
	// will be routed by the kernel to this socket.
	err = syscall.Bind(fd, &syscall.SockaddrInet4{
		Port: s.Port,
		Addr: s.Addr,
	})
	if err != nil {
		return nil, fmt.Errorf("error binding port to tcp socket: %w", err)
	}

	// 3. listen()
	// We switch the socket into "server mode".
	//
	// Now the kernel:
	// - starts accepting incoming TCP connections
	// - performs TCP handshake (SYN → SYN-ACK → ACK)
	// - stores new connections in a queue (backlog queue)
	//
	// backlog  means:
	// "the kernel can keep up to s.Backlog pending connections
	// waiting in the queue before they are accepted."
	//
	// IMPORTANT:
	// This is NOT a limit of total clients.
	// It is only a limit of waiting (unaccepted) connections.
	err = syscall.Listen(fd, s.Backlog)
	if err != nil {
		return nil, fmt.Errorf("error listening tcp socket: %w", err)
	}

	return &Server{
		fd:   fd,
		exec: &executor.Exec{DB: db},
	}, nil
}

func (s *Server) Run() error {
	for {
		// We take one connection from the kernel backlog queue.
		//
		// The kernel:
		// - completes the TCP connection
		// - creates a NEW socket for this client
		// - returns a new file descriptor (nfd)
		//
		// Important idea:
		// fd = listener socket (only accepts connections)
		// nfd = client socket (used for communication)
		nfd, _, err := syscall.Accept(s.fd)
		if err != nil {
			fmt.Printf("error when accepting connection: %s", err.Error())
			continue
		}

		go s.handleConn(nfd)
	}
}

func (s *Server) handleConn(nfd int) {
	buf := make([]byte, 1024)
	for {
		n, err := syscall.Read(nfd, buf)
		if err != nil {
			fmt.Printf("error reading request: %s", err.Error())
			continue
		}

		if n == 0 {
			return
		}

		// TODO there need remove parseCMD from cli package
		// maybe create own module, or explode it to another method without CLI struct
		stmt := cli.ParseCMD(string(buf[:n]))

		err, data := s.exec.ExecStmt(stmt)
		if err != nil {
			syscall.Write(nfd, []byte(err.Error()))
			continue
		}

		if data == nil {
			syscall.Write(nfd, []byte("s"))
			continue
		}

		rows := make([][]any, len(data.Rows))
		for i, v := range data.Rows {
			rows[i] = make([]any, len(v))
			j := 0
			for _, f := range v {
				rows[i][j] = f
				j++
			}
		}

		b, err := json.Marshal(rows)
		if err != nil {
			syscall.Write(nfd, []byte(err.Error()))
			continue
		}

		syscall.Write(nfd, b)
	}
}
