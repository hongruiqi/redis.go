package redis

import (
	"net"
	"sync"
)

/*
 * Client
 */

type Client struct {
	mutex   sync.Mutex
	conn    net.Conn
	pending Commands
	running bool
	enc     *Encoder
	dec     *Decoder
}

func Dial(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, enc: NewEncoder(conn), dec: NewDecoder(conn)}, nil
}

func (c *Client) Execute(cmd *Command) Reply {
	replyChan := c.Go(cmd)
	reply := <-replyChan
	return reply
}

func (c *Client) Go(cmd *Command) <-chan Reply {
	c.mutex.Lock()
	if cmd.ReplyChan == nil {
		cmd.ReplyChan = make(chan Reply, 2) // buffered
	}
	c.pending = append(c.pending, cmd)
	c.mutex.Unlock()
	go c.Serve()
	return cmd.ReplyChan
}

func (c *Client) Serve() {
	for {
		c.mutex.Lock()
		if c.running || len(c.pending) == 0 {
			c.mutex.Unlock()
			return
		}
		c.running = true
		cmd := c.pending[0]
		c.pending = c.pending[1:]
		c.mutex.Unlock()
		err := c.enc.Encode(cmd.Cmd)
		if err != nil {
			cmd.ReplyChan <- &InvalidReply{err}
		}
		reply, err := c.dec.Decode()
		if err != nil {
			cmd.ReplyChan <- &InvalidReply{err}
		} else {
			cmd.ReplyChan <- reply
		}
	}
	c.mutex.Lock()
	c.running = false
	c.mutex.Unlock()
}

/*
 * PipeLines
 */

type PipeLine struct {
	mutex   sync.Mutex
	conn    net.Conn
	cmds    Commands
	pending []Commands
	running bool
	enc     *Encoder
	dec     *Decoder
}

func DialPipeLine(address string) (*PipeLine, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &PipeLine{conn: conn, enc: NewEncoder(conn), dec: NewDecoder(conn)}, nil
}

func (p *PipeLine) Go(cmd *Command) <-chan Reply {
	p.mutex.Lock()
	if cmd.ReplyChan == nil {
		cmd.ReplyChan = make(chan Reply, 2)
	}
	p.cmds = append(p.cmds, cmd)
	p.mutex.Unlock()
	return cmd.ReplyChan
}

func (p *PipeLine) Discard() {
	p.mutex.Lock()
	p.pending = nil
	p.mutex.Unlock()
}

func (p *PipeLine) Exec() {
	p.mutex.Lock()
	p.pending = append(p.pending, p.cmds)
	p.cmds = nil
	p.mutex.Unlock()
	go p.Serve()
}

func (p *PipeLine) Serve() {
Loop:
	for {
		p.mutex.Lock()
		if p.running || len(p.pending) == 0 {
			p.mutex.Unlock()
			return
		}
		p.running = true
		cmds := p.pending[0]
		p.pending = p.pending[1:]
		p.mutex.Unlock()
		var waiting Commands
		for _, cmd := range cmds {
			err := p.enc.Encode(cmd.Cmd)
			if err != nil {
				cmd.ReplyChan <- &InvalidReply{err}
				continue Loop
			}
			waiting = append(waiting, cmd)
		}
		for _, cmd := range waiting {
			reply, err := p.dec.Decode()
			if err != nil {
				cmd.ReplyChan <- &InvalidReply{err}
			} else {
				cmd.ReplyChan <- reply
			}
		}
	}
	p.running = false
	p.mutex.Unlock()
}

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
