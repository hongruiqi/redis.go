package redis

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	w := new(bytes.Buffer)
	encoder := NewEncoder(w)
	err := encoder.Encode([]interface{}{"set", "mykey", 123})
	if err != nil {
		t.Fatal(err)
	}
	got, err := ioutil.ReadAll(w)
	if err != nil {
		t.Fatal(err)
	}
	expect := []byte("*3\r\n$3\r\nset\r\n$5\r\nmykey\r\n$3\r\n123\r\n")
	if !bytes.Equal(got, expect) {
		t.Fatalf("expected %q got %q\n", expect, got)
	}
}

var decodetests = []struct {
	input  string
	output Reply
}{
	{
		"+OK\r\n",
		&StatusReply{"OK"},
	},
	{
		"-Error\r\n",
		&ErrorReply{"Error"},
	},
	{
		":12345\r\n",
		&IntegerReply{12345},
	},
	{
		"$3\r\nkey\r\n",
		&BulkReply{"key"},
	},
}

func TestDecode(t *testing.T) {
	for _, test := range decodetests {
		buf := bytes.NewBufferString(test.input)
		decoder := NewDecoder(buf)
		reply, err := decoder.Decode()
		if err != nil {
			t.Log(err)
			t.Fail()
			continue
		}
		if reflect.TypeOf(reply) != reflect.TypeOf(test.output) {
			t.Logf("decode for %q failed: type error", test.input)
		}
		if reply.Value() != test.output.Value() {
			t.Logf("decode for %q failed: value error", test.input)
			t.Fail()
		}
	}
}

// TODO: MultiBulk decode test
