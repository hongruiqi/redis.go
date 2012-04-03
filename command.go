package redis

type Command struct {
	Cmd       []interface{}
	ReplyChan chan bool
	Reply     Reply
	Error     error
}

type Commands []*Command
