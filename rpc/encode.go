package rpc

import (
	"encoding/binary"
	"net"
)

const lenBytes = 8

func ReadMsg(conn net.Conn) ([]byte, error) {
	lengthBytes := make([]byte, lenBytes)
	_, err := conn.Read(lengthBytes)
	if err != nil {
		return nil, err
	}

	// 解码长度并读取内容
	headLength := binary.BigEndian.Uint32(lengthBytes[:4])
	bodyLength := binary.BigEndian.Uint32(lengthBytes[4:lenBytes])
	data := make([]byte, headLength+bodyLength)
	_, err = conn.Read(data[lenBytes:])
	if err != nil {
		return nil, err
	}
	copy(data[:lenBytes], lengthBytes)

	return data, nil
}
