/*
redis.go is a simple client for redis written in go


Usage

Example:

	import (
		"github.com/hongruiqi/redis.go"
	)

	func testSync() {
		client := redis.client("127.0.0.1:6379")
		cmd := &redis.Command{Cmd: redis.Cmd{"set", "key", "value"}}
		ok := redis.Execute()
		if !ok {
			// error occurred
		} else {
			reply = cmd.Reply
		}
	}

	func testAsync() {
		client := redis.client("127.0.0.1:6379")
		cmd := &redis.Command{Cmd: redis.Cmd{"set", "key", "value"}}
		replyChan := redis.Go(cmd)
		ok := <-replyChan
		if !ok {
			// error occurred
		} else {
			reply = cmd.Reply
		}
	}
*/
package redis
