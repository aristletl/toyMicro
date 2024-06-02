package rpc

import (
	"context"
	"errors"
	"github.com/silenceper/pool"
	"github.com/ztruane/toy-micro/rpc/compress"
	"github.com/ztruane/toy-micro/rpc/message"
	"github.com/ztruane/toy-micro/rpc/serialize"
	"github.com/ztruane/toy-micro/rpc/serialize/json"
	"net"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"
)

var messageId uint32 = 0

type Client struct {
	connPool   pool.Pool
	serializer serialize.Serializer
	compressor compress.Compressor
}

func NewClient(addr string) (*Client, error) {
	pl, err := pool.NewChannelPool(&pool.Config{
		InitialCap: 10,  // 初始化连接容量
		MaxCap:     100, // 最大容量
		MaxIdle:    20,  // 最大空闲连接数量
		Factory: func() (interface{}, error) {
			return net.Dial("tcp", addr)
		},
		IdleTimeout: time.Minute,
		Close: func(i interface{}) error {
			return i.(net.Conn).Close()
		},
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		connPool:   pl,
		serializer: json.Serializer{},
		compressor: compress.DoNothingCompressor{},
	}, nil
}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var resp *message.Response
	var err error
	ch := make(chan struct{})
	go func() {
		resp, err = c.doInvoke(ctx, req)
		close(ch)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ch:
		return resp, err
	}
}

func (c *Client) doInvoke(ctx context.Context, req *message.Request) (*message.Response, error) {

	// 拿一个连接
	obj, err := c.connPool.Get()
	// 这里为框架内部error
	if err != nil {
		return nil, err
	}

	conn, ok := obj.(net.Conn)
	if !ok {
		return nil, errors.New("micro: 非连接错误")
	}
	// 发送数据
	reqMsg := message.EncodeReq(req)
	i, er := conn.Write(reqMsg)
	if er != nil {
		return nil, er
	}
	if i != len(reqMsg) {
		return nil, errors.New("micro: 数据写入发生错误")
	}

	// 读响应，但是应该读多长呢?
	result, err := ReadMsg(conn)
	if err != nil {
		return nil, err
	}

	return message.DecodeResponse(result), nil
}

// 客户端
// 1.首先反射拿到Request， 核心是服务名称，方法名和参数
// 2.将Request进行编码，要注意序列化加上长度
// 3.使用连接池或者一个链接，将请求发过去
// 4.读取响应，解析成结构体

// 服务端
// 1.启动一个服务，监听一个端口
// 2.读取长度字段，再根据长度，读取完整信息
// 3.解析成Request
// 4.查找服务，找到对应的方法
// 5.构造方法对应的输入
// 6.反射执行调用
// 7.编码响应
// 8.写回响应

func (c *Client) InitClientProxy(service Service) error {
	// 这里需要按照约定的格式对传入的service进行校验
	val := reflect.ValueOf(service)
	typ := reflect.TypeOf(service)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}
	numField := val.NumField()
	for i := 0; i < numField; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		if fieldTyp.Type.Kind() != reflect.Func {
			continue
		}

		// 替换为一个新的方法实现
		fn := reflect.MakeFunc(fieldTyp.Type,
			func(args []reflect.Value) (results []reflect.Value) {
				ctx, ok := args[0].Interface().(context.Context)
				// 第一个返回值是真正的响应，需要先获取该接口的响应类型
				outTyp := fieldTyp.Type.Out(0)
				if !ok {
					panic("xxx")
				}
				arg := args[1].Interface()
				bs, er := c.serializer.Encode(arg)
				if er != nil {
					results = append(results, reflect.Zero(outTyp))
					results = append(results, reflect.ValueOf(er))
					return
				}
				bs, er = c.compressor.Compress(bs)
				if er != nil {
					results = append(results, reflect.Zero(outTyp))
					results = append(results, reflect.ValueOf(er))
					return
				}

				meta := make(map[string]string)
				deadline, ok := ctx.Deadline()
				if ok {
					meta["timeout"] = strconv.FormatInt(deadline.Unix(), 10)
				}

				//  在该方法中进行调用信息的拼凑
				// 包括服务名、方法名、参数值
				msgId := atomic.AddUint32(&messageId, 1)
				req := &message.Request{
					Serializer:  c.serializer.Code(),
					Compressor:  c.compressor.Code(),
					MessageID:   msgId,
					BodyLength:  uint32(len(bs)),
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Data:        bs,
					Meta:        meta,
				}

				req.CalHeadLength()

				res, err := c.Invoke(ctx, req)

				if err != nil {
					results = append(results, reflect.Zero(outTyp))
					results = append(results, reflect.ValueOf(err))
					return
				}
				// 创建对应响应的返回值的指针
				out := reflect.New(outTyp.Elem()).Interface()
				data, err := c.compressor.Uncompress(res.Data)
				if err == nil {
					// 这里涉及到rpc协议的数据编码转化
					err = c.serializer.Decode(data, out)
					// 第一个返回值
					results = append(results, reflect.ValueOf(out))
				}
				// 第二个返回值是err
				if err != nil {
					results = append(results, reflect.ValueOf(err))
				} else {
					results = append(results, reflect.Zero(reflect.TypeOf(new(error)).Elem()))
				}
				return
			},
		)

		fieldVal.Set(fn)
	}
	return nil
}
