package message

import (
	"bytes"
	"encoding/binary"
)

const (
	splitter     = '\n'
	pairSplitter = '\r'
)

type Request struct {
	HeadLength uint32
	BodyLength uint32

	MessageID  uint32
	Version    byte
	Compressor byte
	Serializer byte

	ServiceName string
	MethodName  string

	Meta map[string]string
	Data []byte
}

func (r *Request) CalHeadLength() {
	res := 15
	res += len(r.ServiceName)
	res++
	res += len(r.MethodName)
	res++

	for k, v := range r.Meta {
		res = res + len(k) + 1 + len(v) + 1
	}

	r.HeadLength = uint32(res)
}

func EncodeReq(req *Request) []byte {
	bs := make([]byte, req.HeadLength+req.BodyLength)

	cur := bs

	binary.BigEndian.PutUint32(cur[:4], req.HeadLength)
	cur = cur[4:]

	binary.BigEndian.PutUint32(cur[:4], req.BodyLength)
	cur = cur[4:]

	binary.BigEndian.PutUint32(cur[:4], req.MessageID)
	cur = cur[4:]

	cur[0] = req.Version
	cur = cur[1:]
	cur[0] = req.Compressor
	cur = cur[1:]
	cur[0] = req.Serializer
	cur = cur[1:]

	copy(cur, req.ServiceName)
	// 加个分隔符
	cur[len(req.ServiceName)] = splitter
	cur = cur[len(req.ServiceName)+1:]

	copy(cur, req.MethodName)
	// 加个分隔符
	cur[len(req.MethodName)] = splitter
	cur = cur[len(req.MethodName)+1:]

	for key, value := range req.Meta {
		copy(cur, key)
		cur[len(key)] = pairSplitter
		cur = cur[len(key)+1:]

		copy(cur, value)
		cur[len(value)] = splitter
		cur = cur[len(value)+1:]
	}

	copy(cur, req.Data)

	return bs
}

func DecodeReq(data []byte) *Request {
	req := new(Request)

	req.HeadLength = binary.BigEndian.Uint32(data[:4])
	req.BodyLength = binary.BigEndian.Uint32(data[4:8])
	req.MessageID = binary.BigEndian.Uint32(data[8:12])
	req.Version = data[12]
	req.Compressor = data[13]
	req.Serializer = data[14]

	// 剩余为头部数据
	head := data[15:req.HeadLength]
	index := bytes.IndexByte(head, splitter)
	req.ServiceName = string(head[:index])

	head = head[index+1:]
	index = bytes.IndexByte(head, splitter)
	req.MethodName = string(head[:index])

	head = head[index+1:]
	if len(head) > 0 {
		meta := make(map[string]string)
		index = bytes.IndexByte(head, splitter)
		for index != -1 {
			pair := head[:index]
			pairIndex := bytes.IndexByte(pair, pairSplitter)
			key := string(pair[:pairIndex])
			value := string(pair[pairIndex+1:])
			meta[key] = value

			head = head[index+1:]
			index = bytes.IndexByte(head, splitter)
		}
		req.Meta = meta
	}

	req.Data = data[req.HeadLength:]

	return req
}

type Response struct {
	HeadLength uint32
	BodyLength uint32

	MessageID  uint32
	Version    byte
	Compressor byte
	Serializer byte

	Error    string
	BizError string // 区分业务错误还是内部错误

	Data []byte
}

func (resp *Response) CalHeadLength() {
	resp.HeadLength = uint32(15 + len(resp.Error))
}

func EncodeResp(resp *Response) []byte {
	bs := make([]byte, resp.HeadLength+resp.BodyLength)

	cur := bs

	binary.BigEndian.PutUint32(cur[:4], resp.HeadLength)
	cur = cur[4:]

	binary.BigEndian.PutUint32(cur[:4], resp.BodyLength)
	cur = cur[4:]

	binary.BigEndian.PutUint32(cur[:4], resp.MessageID)
	cur = cur[4:]

	cur[0] = resp.Version
	cur = cur[1:]
	cur[0] = resp.Compressor
	cur = cur[1:]
	cur[0] = resp.Serializer
	cur = cur[1:]

	copy(cur, resp.Error)
	cur = cur[len(resp.Error):]

	copy(cur, resp.Data)

	return bs
}

func DecodeResponse(data []byte) *Response {
	resp := new(Response)

	resp.HeadLength = binary.BigEndian.Uint32(data[:4])
	resp.BodyLength = binary.BigEndian.Uint32(data[4:8])
	resp.MessageID = binary.BigEndian.Uint32(data[8:12])
	resp.Version = data[12]
	resp.Compressor = data[13]
	resp.Serializer = data[14]

	// 剩余为错误信息数据
	resp.Error = string(data[15:resp.HeadLength])

	resp.Data = data[resp.HeadLength:]

	return resp
}
