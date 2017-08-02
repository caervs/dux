package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/caervs/dux/api"
)

const (
	target = "golang.org:80"
)

type DuxDaemon struct {
	connections map[int64]io.ReadWriteCloser
	in          io.Reader
	out         *os.File
	lock        *sync.Mutex
	logger      *log.Logger
}

func NewStandardDuxDaemon() *DuxDaemon {
	var buf bytes.Buffer
	return &DuxDaemon{
		connections: make(map[int64]io.ReadWriteCloser),
		in:          os.Stdin,
		out:         os.Stdout,
		lock:        &sync.Mutex{},
		logger:      log.New(&buf, "logger: ", log.Lshortfile),
	}
}

func (dd *DuxDaemon) read(session int64, conn io.ReadCloser) {
	// TODO capture errors
	buffer := make([]byte, 1024)
	header := &api.Header{
		Session: session,
	}
	for {
		n, err := conn.Read(buffer)
		if err == io.EOF {
			header.Size = 0
			dd.lock.Lock()
			// TODO delete the connection
			_, err = dd.out.Write(header.ToBytes())
			dd.lock.Unlock()
			if err != nil {
				log.Print("error closing ", err)
			}
			return
		} else if err != nil {
			log.Print("error while reading ", err)
			return
		}
		header.Size = n
		dd.lock.Lock()
		_, err1 := dd.out.Write(header.ToBytes())
		// TODO slicing is not super efficient here
		_, err2 := dd.out.Write(buffer[:n])
		dd.lock.Unlock()
		if err1 != nil || err2 != nil {
			log.Print("error writing", err1, err2)
		}
	}
}

func (dd *DuxDaemon) next() error {
	// TODO track errors so they can be written back to client
	header, err := api.ReadHeader(dd.in)
	if err != nil {
		return err
	}
	log.Print("Server got header for session ", header.Session, " with size ", header.Size)
	conn, ok := dd.connections[header.Session]
	if !ok {
		conn, err = net.Dial("tcp", target)
		if err != nil {
			return err
		}
		dd.connections[header.Session] = conn
		go dd.read(header.Session, conn)
	}
	if header.Size == 0 {
		// TODO fix
		//err = conn.Close()
		//if err != nil {
		//	return err
		//}
		//delete(dd.connections, header.Session)
	}
	// TODO consider reusing buffer if performance is bad
	buffer := make([]byte, header.Size)
	_, err = io.ReadFull(dd.in, buffer)
	if err != nil {
		return err
	}
	_, err = conn.Write(buffer)
	if err != nil {
		return err
	}
	return nil
}

func (dd *DuxDaemon) Serve() {
	for {
		err := dd.next()
		if err != nil {
			log.Print("error getting next ", err)
		}
	}
}

func main() {
	dd := NewStandardDuxDaemon()
	dd.Serve()
}
