package proto

import (
	"errors"
	"google.golang.org/protobuf/proto"
)

type Serializer struct {
}

func (s Serializer) Code() byte {
	return 0
}

func (s Serializer) Encode(val any) ([]byte, error) {
	msg, ok := val.(proto.Message)
	if !ok {
		return nil, errors.New("hahah")
	}
	return proto.Marshal(msg)
}

func (s Serializer) Decode(bytes []byte, val any) error {
	msg, ok := val.(proto.Message)
	if !ok {
		return errors.New("")
	}
	return proto.Unmarshal(bytes, msg)
}
