package redis

import (
	"net"
	"sync"
)

/*
 * PubSub
 */

type PubSub struct {
	mutex   sync.Mutex
	conn    net.Conn
	MsgChan chan Reply
	Quit    chan bool
	enc     *Encoder
	dec     *Decoder
}

func DialPubSub(address string) (*PubSub, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	ps := &PubSub{conn: conn, enc: NewEncoder(conn), dec: NewDecoder(conn), MsgChan: make(chan Reply), Quit: make(chan bool)}
	go ps.listen()
	return ps, nil
}

func (ps *PubSub) Subscribe(channel ...interface{}) error {
	return ps.subscribe("subscribe", channel...)
}

func (ps *PubSub) PSubscribe(channel ...interface{}) error {
	return ps.subscribe("psubscribe", channel...)
}

func (ps *PubSub) subscribe(t string, channel ...interface{}) error {
	var cmd []interface{}
	cmd = append(cmd, t)
	cmd = append(cmd, channel...)
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	err := ps.enc.Encode(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (ps *PubSub) listen() error {
	for {
		select {
		case <-ps.Quit:
			return nil
		default:
		}
		reply, err := ps.dec.Decode()
		if err != nil {
			ps.MsgChan <- &InvalidReply{err}
		} else {
			ps.MsgChan <- reply
		}

	}
	return nil
}
