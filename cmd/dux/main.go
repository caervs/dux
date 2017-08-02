package main

import (
	"bufio"
	"io"
	"log"
	"math/rand"
	"net"
	"os/exec"
	"sync"

	"github.com/caervs/dux/api"
)

type DuxClient struct {
	out         io.WriteCloser
	in          io.ReadCloser
	lock        *sync.Mutex
	connections map[int64]io.ReadWriteCloser
}

func NewDuxClient(out io.WriteCloser, in io.ReadCloser) *DuxClient {
	return &DuxClient{
		out:         out,
		in:          in,
		lock:        &sync.Mutex{},
		connections: make(map[int64]io.ReadWriteCloser),
	}
}

func (dc *DuxClient) Handle(conn net.Conn) {
	session := rand.Int63()
	dc.connections[session] = conn
	buffer := make([]byte, 1024)
	header := &api.Header{
		Session: session,
	}
	// TODO refactor with daemon
	for {
		n, err := conn.Read(buffer)
		if err == io.EOF {
			header.Size = 0
			dc.lock.Lock()
			// TODO delete the connection
			_, err = dc.out.Write(header.ToBytes())
			dc.lock.Unlock()
			if err != nil {
				log.Print("error closing", err)
			}
			return
		} else if err != nil {
			log.Print("error while reading", err)
			return
		}
		header.Size = n
		dc.lock.Lock()
		_, err1 := dc.out.Write(header.ToBytes())
		// TODO slicing is not super efficient here
		_, err2 := dc.out.Write(buffer[:n])
		log.Print(buffer[:n])
		dc.lock.Unlock()
		if err1 != nil || err2 != nil {
			log.Print("error writing ", err1, err2)
		}
		log.Print("Client wrote header for session ", header.Session, " with size ", header.Size)
	}
}

func (dc *DuxClient) Read() {
	for {
		log.Print("asking for header")
		header, err := api.ReadHeader(dc.in)
		if err != nil {
			log.Print("error reading ", err)
		}
		log.Print("Client got header with session ", header.Session, "and size", header.Size)
		// TODO consider reusing buffer if performance is bad
		buffer := make([]byte, header.Size)
		n, err := io.ReadFull(dc.in, buffer)
		if err != nil {
			log.Print("error reading ", err)
		}
		conn, ok := dc.connections[header.Session]
		if !ok {
			log.Print("Ignoring unknown session", header.Session)
		}
		if n == 0 {
			log.Print("Closing session ", header.Session)
			// HACK
			conn.Write([]byte("\r\n"))
			err = conn.Close()
			if err != nil {
				log.Print("error closing", err)
			}
		}
		_, err = conn.Write(buffer)
		log.Print("buffer is ", string(buffer))
		if err != nil {
			log.Print("error writing", err)
		}
	}
}

// TODO should be in docker
func startd() (io.WriteCloser, io.ReadCloser, error) {
	cmd := exec.Command("./duxd")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	go func() {
		r := bufio.NewReader(stderr)
		for {
			s, err := r.ReadString('\n')
			if err != nil {
				log.Print("Got error")
				return
			}
			log.Print("Server: ", s)
		}
	}()
	err = cmd.Start()
	if err != nil {
		return nil, nil, err
	}
	return stdin, stdout, err
}

func main() {
	stdin, stdout, err := startd()
	if err != nil {
		panic(err)
	}
	dc := NewDuxClient(stdin, stdout)
	go dc.Read()
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go dc.Handle(conn)
	}
}
