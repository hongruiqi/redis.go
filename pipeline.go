package redis

import (
	"net"
	"sync"
)

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

func (p *PipeLine) Go(cmd *Command) <-chan bool {
	p.mutex.Lock()
	if cmd.ReplyChan == nil {
		cmd.ReplyChan = make(chan bool, 2)
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
				cmd.Error = err
				cmd.ReplyChan <- false
				continue
			}
			waiting = append(waiting, cmd)
		}
		for _, cmd := range waiting {
			reply, err := p.dec.Decode()
			if err != nil {
				cmd.Error = err
				cmd.ReplyChan <- false
			} else {
				cmd.Reply = reply
				cmd.ReplyChan <- true
			}
		}
	}
	p.mutex.Lock()
	p.running = false
	p.mutex.Unlock()
}
