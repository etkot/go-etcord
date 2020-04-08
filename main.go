package main

import (
	"bytes"
	"net"
	"os"
	"os/signal"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	mu sync.Mutex
	wg sync.WaitGroup

	stop chan struct{}
	port string
	clients map[string]net.Conn // TODO
}

func NewServer(port string) *Server {
	return &Server{
		port: port,
		stop: make(chan struct{}),
		clients: make(map[string]net.Conn),
	}
}

func (s *Server) Start() {
	s.wg.Add(1)
	go s.tcpServer()

	log.Info("Starting Etcord server")
}

func (s *Server) Stop() {
	log.Info("Stopping Etcord server")
	close(s.stop)
}

func (s *Server) tcpServer() {
	defer s.wg.Done()

	l, err := net.Listen("tcp4", ":"+s.port)
	if err != nil {
		log.Errorf("Failed to start listener: %s", err)
		return
	}
	defer l.Close()

	log.Infof("Server listening on %s", s.port)

	go func() {
		select {
		case <-s.stop:
			l.Close()
			return
		}
	}()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Errorf("Failed to accept connection: %s", err)
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			break
		}
		go s.handleConn(c)
	}
}

func (s *Server) handleConn(c net.Conn) {
	log.Infof("Connected to %s", c.RemoteAddr().String())

	s.addClient(c)
	defer s.removeClient(c)

	tmp := make([]byte, 1024)
	for {
		n, err := c.Read(tmp)
		if err != nil {
			log.Error(err)
			break
		}
		log.Debugf("Read %d bytes: [%x]", n, tmp[:n])

		b := bytes.NewBuffer(tmp[:n])
		var msgs []*Msg
		for {
			if b.Len() == 0 {
				break
			}
			m := NewMsg()
			if err := m.Deserialize(b); err != nil {
				log.Errorf("Failed to deserialize msg from buffer: %s", err)
				break
			}
			msgs = append(msgs, m)
		}

		for _, m := range msgs {
			s.msgHandler(m)
		}
	}

	if err := c.Close(); err != nil {
		log.Errorf("Failed to close connection: %s", err)
	}
}

func (s *Server) msgHandler(m *Msg) {
	log.Infof("Recv msg: %s", m)
}

func (s *Server) addClient(c net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[c.RemoteAddr().String()] = c
}

func (s *Server) removeClient(c net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, c.RemoteAddr().String())
}

func main() {
	if len(os.Args) != 2 {
		log.Error("Missing port number")
		return
	}

	s := NewServer(os.Args[1])
	s.Start()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		select {
		case <-sig:
			s.Stop()
		}

		for {
			select {
			case <-sig:
				log.Info("Already shutting down")
			}
		}
	}()

	s.wg.Wait()
}

func init() {
	log.SetLevel(log.TraceLevel)
}
