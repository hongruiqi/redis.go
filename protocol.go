package redis

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

/*
 * Encoder
 */

func Encode(v []interface{}) ([]byte, error) {
	buf := bytes.NewBufferString("")
	count := len(v)
	_, err := fmt.Fprintf(buf, "*%d\r\n", count)
	if err != nil {
		return nil, err
	}
	for _, i := range v {
		s := fmt.Sprint(i)
		n := len(s)
		_, err = fmt.Fprintf(buf, "$%d\r\n%s\r\n", n, s)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (enc *Encoder) Encode(v []interface{}) error {
	e, err := Encode(v)
	if err != nil {
		return err
	}
	_, err = enc.w.Write(e)
	if err != nil {
		return err
	}
	return nil
}

/*
 * Decoder
 */

type Decoder struct {
	r   io.Reader
	buf *bufio.Reader
	err error
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r, buf: bufio.NewReader(r)}
}

func (dec *Decoder) Decode() (Reply, error) {
	if dec.err != nil {
		return nil, dec.err
	}
	reply, err := dec.decode()
	if err != nil {
		dec.err = err
		return nil, err
	}
	switch reply.(type) {
	case *StatusReply, *ErrorReply, *IntegerReply, *BulkReply:
		return reply, nil
	case *MultiBulkReply:
		mb := reply.(*MultiBulkReply)
		for count := reply.(*MultiBulkReply).count; count > 0; count-- {
			reply, err := dec.decode()
			if err != nil {
				dec.err = err
				return nil, err
			}
			mb.Append(reply)
		}
		return reply, nil
	}
	panic("should not reach here")
	return nil, nil
}

func (dec *Decoder) decode() (Reply, error) {
	s, isPrefix, err := dec.buf.ReadLine()
	if isPrefix {
		return nil, errors.New("reply line too long")
	}
	if err != nil {
		return nil, err
	}
	msg := string(s[1:])
	switch s[0] {
	case '+':
		return &StatusReply{msg}, nil
	case '-':
		return &ErrorReply{msg}, nil
	case ':':
		num, err := strconv.ParseInt(msg, 10, 64)
		if err != nil {
			return nil, errors.New("invalid IntegerReply")
		}
		return &IntegerReply{num}, nil
	case '$':
		bulkLen, err := strconv.ParseInt(msg, 10, 64)
		if err != nil {
			return nil, errors.New("invalid bulk length")
		}
		b := make([]byte, bulkLen+2)
		_, err = io.ReadFull(dec.buf, b)
		if err != nil {
			return nil, err
		}
		return &BulkReply{string(b[:bulkLen])}, nil
	case '*':
		bulksCount, err := strconv.ParseInt(msg, 10, 64)
		if err != nil {
			return nil, errors.New("invalid multibulk count")
		}
		return NewMultiBulkReply(bulksCount), nil
	default:
		return nil, errors.New("invalid reply type")
	}
	panic("should not reach here")
	return nil, nil
}
