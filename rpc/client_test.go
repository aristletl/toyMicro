package rpc

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ztruane/toy-micro/rpc/message"
	"testing"
)

func TestInitClientProxy(t *testing.T) {
	testCases := []struct {
		name string

		service *UserService
		proxy   *mockProxy

		wantErr error
		wantReq *message.Request
	}{
		{
			name:    "",
			service: &UserService{},
			//proxy:    ,

			wantReq: &message.Request{
				ServiceName: "user-service",
				MethodName:  "GetById",
				Data:        []byte(`{"Id":123}`),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := NewClient(":8081")
			err := p.InitClientProxy(tc.service)
			assert.Equal(t, tc.wantErr, err)
			resp, err := tc.service.GetById(context.Background(), &GetByIdReq{Id: 123})
			require.Nil(t, err)
			fmt.Println(resp)
			//assert.Equal(t, tc.wantReq, tc.proxy.message.Request)
			//assert.Equal(t, &GetByIdResp{Name: "abc"}, resp)
		})
	}
}

type GetByIdReq struct {
	Id int
}

type GetByIdResp struct {
	Name string
}

type UserService struct {
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
}

func (u *UserService) Name() string {
	return "user-service"
}

type mockProxy struct {
	request *message.Request
	data    []byte
}

func (m *mockProxy) Invoke(ctx context.Context, req *message.Request) (message.Response, error) {
	m.request = req
	return message.Response{Data: m.data}, nil
}

func TestA(t *testing.T) {
	ans := permute([]int{1, 2, 3})
	fmt.Println(ans)
}

func permute(nums []int) [][]int {
	var ans [][]int
	var choosen []int
	var flag = make([]bool, len(nums))
	var handle func(int)

	handle = func(i int) {

		choosen = append(choosen, nums[i])
		flag[i] = true
		if len(choosen) == len(nums) {
			ans = append(ans, append([]int{}, choosen...))
			return
		}
		for j := 0; j < len(nums); j++ {
			if flag[j] {
				continue
			}
			handle(j)
		}
		choosen = choosen[:len(choosen)-1]
		flag[i] = false
	}

	for i := 0; i < len(nums); i++ {
		handle(i)
	}
	return ans

}
