package redis

type Cmd []interface{}

type Command struct {
	Cmd       Cmd
	replyChan chan bool
	Reply     Reply
	Error     error
}

type Commands []*Command
