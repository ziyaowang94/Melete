package p2p

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Sender struct {
	shardMap    map[string][]*Peer
	connMapLock sync.Mutex

	connMap map[string]net.Conn
	lockMap map[string]*sync.Mutex

	RetryDuration time.Duration

	MyIP string
}

func NewSender(ip string) *Sender {
	return &Sender{
		shardMap:    map[string][]*Peer{},
		connMapLock: sync.Mutex{},
		connMap:     map[string]net.Conn{},
		lockMap:     map[string]*sync.Mutex{},

		RetryDuration: 100 * time.Millisecond,
		MyIP:          ip,
	}
}

func (s *Sender) Start() error {
	errList := []string{}
	log.Println("My broadcast list", s.shardMap)
	s.connMapLock.Lock()
	defer s.connMapLock.Unlock()
	for _, pl := range s.shardMap {
		for _, p := range pl {
			if p.GetIP() != s.MyIP && s.connMap[p.GetIP()] == nil {
				if conn, err := net.Dial("tcp", p.GetIP()); err != nil {
					errList = append(errList, err.Error())
				} else {
					if err := conn.SetDeadline(time.Time{}); err != nil {
						panic(err)
					}
					s.connMap[p.GetIP()] = conn
					fmt.Println(time.Now(), "--node", p.GetIP(), "add success")
				}
			}
		}
	}
	if len(errList) == 0 {
		return nil
	} else {
		return fmt.Errorf(strings.Join(errList, "\n"))
	}
}

func (s *Sender) Stop() error {
	for _, conn := range s.connMap {
		conn.Close()
	}
	return nil
}

func (s *Sender) AddPeer(peer *Peer) {
	for cid := range peer.ChainID() {
		if len(s.shardMap[cid]) == 0 {
			s.shardMap[cid] = []*Peer{peer}
		} else {
			flg := true
			for _, p := range s.shardMap[cid] {
				if peer.Equal(p) {
					flg = false
					break
				}
			}
			if flg {
				s.shardMap[cid] = append(s.shardMap[cid], peer)
			}
		}
	}
}

func (s *Sender) Send(peer *Peer, channel_id byte, message []byte, messageType uint32) error {
	if s.MyIP == peer.GetIP() {
		return nil
	}
	e := Envelop{Channel_id: channel_id, Message: message, MessageType: messageType}
	bz, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	//log.Println(bz)

	return s.TcpDial(bz, peer.GetIP(), 5)
}

func (s *Sender) SendToShard(shardID string, channel_id byte, message []byte, messageType uint32) error {
	peers := s.shardMap[shardID]
	for _, peer := range peers {
		go s.Send(peer, channel_id, message, messageType)
	}
	return nil
}


func (s *Sender) TcpDial(context []byte, addr string, depth int) error {
	var conn net.Conn
	var err error
	s.connMapLock.Lock()
	conn = s.connMap[addr]
	if conn == nil {
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			s.connMapLock.Unlock()
			return err
		}
		if err := conn.SetDeadline(time.Now().Add(30 * time.Minute)); err != nil {
			s.connMapLock.Unlock()
			panic(err)
		}
		fmt.Println(time.Now(), "--node", addr, "add success")
		s.connMap[addr] = conn
	}
	s.connMapLock.Unlock()

	_, err = conn.Write(append(context, '\n'))
	//time.Sleep(time.Duration(len(context)/1024+1) * time.Microsecond)
	if err == nil {
		return nil
	}
	if depth <= 0 {
		return err
	}

	time.Sleep(s.RetryDuration)
	return s.TcpDial(context, addr, depth-1)
}
