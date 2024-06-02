package rpc

import (
	"context"
	"fmt"
	"github.com/ztruane/toy-micro/rpc/message"
	"github.com/ztruane/toy-micro/rpc/serialize"
	"log"
	"net"
	"reflect"
	"strconv"
	"time"
)

type Server struct {
	services    map[string]reflectionSub
	serializers map[byte]serialize.Serializer
}

func NewServer() *Server {
	return &Server{
		services:    make(map[string]reflectionSub),
		serializers: make(map[byte]serialize.Serializer),
	}
}

func (s *Server) Register(service Service) error {
	s.services[service.Name()] = reflectionSub{
		value: reflect.ValueOf(service),
	}
	return nil
}

func (s *Server) MustRegister(serializer serialize.Serializer) {
	s.serializers[serializer.Code()] = serializer
}

func (s *Server) Start(addr string) {
	lister, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	for {
		conn, er := lister.Accept()
		if er != nil {
			log.Println(er)
			continue
		}

		go func() {
			er := s.handleConn(conn)
			if er != nil {
				conn.Close()
				log.Println(er)
			}
		}()

	}

}

func (s *Server) handleConn(conn net.Conn) error {
	for {
		// 先读长度
		data, err := ReadMsg(conn)
		if err != nil {
			return err
		}
		req := message.DecodeReq(data)
		resp := &message.Response{
			MessageID:  req.MessageID,
			Version:    req.Version,
			Compressor: req.Compressor,
			Serializer: req.Serializer,
		}

		serializer, exist := s.serializers[req.Serializer]
		if !exist {
			resp.Error = fmt.Sprint("不支持的序列化协议")
			resp.CalHeadLength()
			_, err = conn.Write(message.EncodeResp(resp))
			if err != nil {
				return err
			}
			continue
		}

		service, ok := s.services[req.ServiceName]
		if !ok {
			resp.Error = fmt.Sprintf("未找到该服务: %s", req.ServiceName)
		} else {
			ctx := context.Background()
			var cancel = func() {}
			if deadline, ok := req.Meta["timeout"]; ok {
				target, e := strconv.ParseInt(deadline, 10, 64)
				if e != nil {
				}
				ctx, cancel = context.WithDeadline(ctx, time.UnixMilli(target))
			}
			result, err := service.Invoke(context.Background(), serializer, req.MethodName, req.Data)
			cancel()
			if err != nil {
				resp.Error = err.Error()
			}
			resp.BodyLength = uint32(len(result))
			resp.Data = result
		}
		resp.CalHeadLength()

		_, err = conn.Write(message.EncodeResp(resp))
		if err != nil {
			return err
		}
		//log.Println(resp)
	}
}

type reflectionSub struct {
	value      reflect.Value
	serializer serialize.Serializer
}

func (r *reflectionSub) Invoke(ctx context.Context, serializer serialize.Serializer, methodName string, data []byte) ([]byte, error) {
	method := r.value.MethodByName(methodName)
	inTyp := method.Type().In(1)
	in := reflect.New(inTyp.Elem())
	err := serializer.Decode(data, in.Interface())
	if err != nil {
		return nil, err
	}

	res := method.Call([]reflect.Value{reflect.ValueOf(ctx), in})
	if len(res) > 1 && !res[1].IsZero() {
		return nil, res[1].Interface().(error)
	}

	return serializer.Encode(res[0].Interface())
}
