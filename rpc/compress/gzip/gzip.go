package gzip

import (
	"compress/gzip"
	"go.uber.org/zap/buffer"
)

type Compressor struct {
}

func (c Compressor) Code() byte {
	//TODO implement me
	panic("implement me")
}

func (c Compressor) Compress(data []byte) ([]byte, error) {
	res := &buffer.Buffer{}
	gw := gzip.NewWriter(res)
	_, err := gw.Write(data)
	if err != nil {
		return nil, err
	}

	err = gw.Close()
	if err != nil {
		return nil, err
	}

	return res.Bytes(), nil
}

func (c Compressor) Uncompress(data []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
