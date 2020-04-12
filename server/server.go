package server

import (
	"etcord/types"
	"fmt"
	"net"
	"sync"

	"etcord/common"
	"etcord/msg"

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
	out chan msg.Msg // outgoing messages
}

func NewServer(port string) *Server {
	return &Server{
		port:     port,
		stop:     make(chan struct{}),
		clients:  make(map[string]*Client),
		channels: make(map[uint16]*Channel),
		out: make(chan msg.Msg),
	}
}

type Client struct {
	UserID uint16 `json:"userId"`
	Name   string `json:"name"`

	conn net.Conn
}

type Channel struct {
	ID uint16      `json:"channelId"`
	ParentID  uint16      `json:"parentId"`
	Name      string      `json:"name"`
	Type      types.ChannelType `json:"type"`

	mu            sync.RWMutex
	lastMessageID int
	messages      map[uint16]*types.Message
}

func NewChannel(channelType types.ChannelType) *Channel {
	// TODO
	return &Channel{
		ID: 0,
		ParentID:  0,
		Name:      "txt",
		Type:      channelType,
		messages:  make(map[uint16]*types.Message),
	}
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

// XXX
func (s *Server) Wait() {
	s.wg.Wait()
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

func (s *Server) AddChannel() *Channel {
	s.mu.Lock()
	defer s.mu.Unlock()
	chn := NewChannel(types.TextChannelType)
	s.channels[chn.ID] = chn
	return chn
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
					b, err := msg.Serialize(m)
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
		var msgs []msg.Msg
		for {
			if buf.Len() == 0 {
				break
			}
			var m msg.Msg
			if m, err = msg.Deserialize(buf); err != nil {
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

func (s *Server) msgHandler(m msg.Msg) error {
	log.Debugf("Recv %s", msg.Type(msg.GetPacketID(m)))

	var err error
	switch m.(type) {
	case *msg.ChatMessageRequest:
		err = s.handleChatMessage(m.(*msg.ChatMessageRequest))
	}

	return err
}

func (s *Server) handleChatMessage(req *msg.ChatMessageRequest) error {
	chn, ok := s.channels[req.ChannelID]
	if !ok {
		return fmt.Errorf("channel with ID %d does not exist", req.ChannelID)
	}

	if chn.Type != types.TextChannelType {
		return fmt.Errorf("channel type is wrong")
	}

	chn.mu.Lock()
	defer chn.mu.Unlock()

	chn.lastMessageID++
	m := &types.Message{
		MessageID: uint16(chn.lastMessageID),
		Content:   req.Content,
	}
	chn.messages[m.MessageID] = m

	res := &msg.ChatMessageResponse{
		ChannelID: chn.ID,
		Message: *m,
	}

	s.out <- msg.Msg(res)

	return nil
}