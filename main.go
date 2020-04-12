package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"

	"etcord/common"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	mu sync.Mutex
	wg sync.WaitGroup

	stop     chan struct{}
	port     string
	lastClientID int
	clients  map[string]*Client
	channels map[uint16]*Channel
	out chan Msg // outgoing messages
}

func NewServer(port string) *Server {
	return &Server{
		port:     port,
		stop:     make(chan struct{}),
		clients:  make(map[string]*Client),
		channels: make(map[uint16]*Channel),
	}
}

type Client struct {
	UserID uint16 `json:"userId"`
	Name   string `json:"name"`

	conn net.Conn
}

func (s *Server) NewClient(conn net.Conn) *Client {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := &Client{
		UserID: uint16(s.lastClientID),
		Name: "teme", // TODO
		conn: conn,
	}
	s.lastClientID++

	return c
}

type ChannelType uint8

const (
	NoneChannelType ChannelType = iota // TODO this would be a category
	TextChannelType
	VoiceChannelType
	MultiChannelType // text + voice
)

type Channel struct {
	ID uint16      `json:"channelId"`
	ParentID  uint16      `json:"parentId"`
	Name      string      `json:"name"`
	Type      ChannelType `json:"type"`

	mu            sync.RWMutex
	lastMessageID int
	messages      map[uint16]*Message
}

func NewChannel(channelType ChannelType) *Channel {
	// TODO
	return &Channel{
		ID: 0,
		ParentID:  0,
		Name:      "txt",
		Type:      channelType,
		messages:  make(map[uint16]*Message),
	}
}

type Message struct {
	MessageID  uint16 `json:"messageId"`
	SenderID   uint16 `json:"senderId"`
	SenderName string `json:"senderName"`
	Content    string `json:"content"`
}

func (s *Server) Start() {
	s.wg.Add(3) // XXX TODO
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
		defer s.wg.Done()
		select {
		case <-s.stop:
			l.Close()
			return
		}
	}()

	// TODO
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.stop:
				return
			case m := <-s.out:
				for _, client := range s.clients {
					b, err := Serialize(m)
					if err != nil {
						log.Errorf("Failed to serialize outgoing message: %s", err)
						break
					}
					if _, err := client.conn.Write(b); err != nil {
						s.removeClient(client)
					}
				}
			}
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

func (s *Server) handleConn(conn net.Conn) {
	log.Infof("Connected to %s", conn.RemoteAddr().String())

	c := s.NewClient(conn)
	s.addClient(c)
	defer s.removeClient(c)

	tmp := make([]byte, 1024)
	for {
		n, err := conn.Read(tmp)
		if err != nil {
			log.Error(err)
			break
		}
		log.Debugf("Read %d bytes: [% x]", n, tmp[:n])

		buf := common.NewBuffer(tmp[:n])
		var msgs []Msg
		for {
			if buf.Len() == 0 {
				break
			}
			var m Msg
			if m, err = Deserialize(buf); err != nil {
				log.Errorf("Failed to deserialize msg from buffer: %s", err)
				break
			}
			msgs = append(msgs, m)
		}

		for _, m := range msgs {
			// TODO error handling
			if err := s.msgHandler(m); err != nil {
				log.Errorf("Failed to process msg: %s", err)
			}
		}
	}

	if err := conn.Close(); err != nil {
		log.Errorf("Failed to close connection: %s", err)
	}

	log.Infof("Disconnected from %s", conn.RemoteAddr().String())
}

func (s *Server) addClient(c *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[c.conn.RemoteAddr().String()] = c
}

func (s *Server) removeClient(c *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, c.conn.RemoteAddr().String())
}

func (s *Server) msgHandler(m Msg) error {
	log.Debugf("Recv msg: %s", m)

	var err error
	switch m.(type) {
	case *ChatMessageRequest:
		err = s.handleChatMessage(m.(*ChatMessageRequest))
	}

	return err
}

func (s *Server) handleChatMessage(req *ChatMessageRequest) error {
	chn, ok := s.channels[req.ChannelID]
	if !ok {
		return fmt.Errorf("channel with ID %d does not exist", req.ChannelID)
	}

	chn.mu.Lock()
	defer chn.mu.Unlock()

	chn.lastMessageID++
	m := &Message{
		MessageID: uint16(chn.lastMessageID),
		Content:   req.Content,
	}
	chn.messages[m.MessageID] = m

	res := &ChatMessageResponse{
		ChannelID: chn.ID,
		Message: *m,
	}

	s.out <- Msg(res)

	return nil
}

func main() {
	// TODO config
	if len(os.Args) != 2 {
		log.Error("Missing port number")
		return
	}
	log.SetLevel(log.TraceLevel)

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
