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

func (c *Client) Execute(cmd *Command) (ok bool) {
	replyChan := c.Go(cmd)
	ok = <-replyChan
	return
}

func (c *Client) Go(cmd *Command) <-chan bool {
	c.mutex.Lock()
	if cmd.replyChan == nil {
		cmd.replyChan = make(chan bool, 2) // buffered
	}
	c.pending = append(c.pending, cmd)
	c.mutex.Unlock()
	go c.Serve()
	return cmd.replyChan
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
			cmd.Error = err
			cmd.replyChan <- false
		}
		reply, err := c.dec.Decode()
		if err != nil {
			cmd.Error = err
			cmd.replyChan <- false
		} else {
			cmd.Reply = reply
			cmd.replyChan <- true
		}
	}
	c.mutex.Lock()
	c.running = false
	c.mutex.Unlock()
}
