package redis

type Reply interface {
	Value() interface{}
}

type InvalidReply struct {
	value error
}

func (r *InvalidReply) Value() interface{} {
	return r.value
}

type StatusReply struct {
	value string
}

func (r *StatusReply) Value() interface{} {
	return r.value
}

type ErrorReply struct {
	value string
}

func (r *ErrorReply) Value() interface{} {
	return r.value
}

type BulkReply struct {
	value string
}

func (r *BulkReply) Value() interface{} {
	return r.value
}

type IntegerReply struct {
	value int64
}

func (r *IntegerReply) Value() interface{} {
	return r.value
}

type MultiBulkReply struct {
	value []Reply
	count int64
}

func NewMultiBulkReply(count int64) *MultiBulkReply {
	return &MultiBulkReply{count: count}
}

func (r *MultiBulkReply) Value() interface{} {
	return r.value
}

func (r *MultiBulkReply) Append(v Reply) {
	r.value = append(r.value, v)
}
