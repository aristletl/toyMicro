package toy_micro

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/ztruane/toy-micro/registry"
	"github.com/ztruane/toy-micro/registry/mocks"
	"google.golang.org/grpc/resolver"
	"testing"
)

func Test_grpcResolverBuilder_Build(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}

func Test_grpcResolver_ResolveNow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCase := []struct {
		name string

		mock func() registry.Registry

		wantState resolver.State
		wantErr   error
	}{
		{
			name: "resolver",

			mock: func() registry.Registry {
				r := mocks.NewMockRegistry(ctrl)
				r.EXPECT().ListerServer(gomock.Any(), gomock.Any()).Return([]registry.ServiceInstance{
					{
						Addr: "test-1",
					},
				}, nil)

				return r
			},

			wantState: resolver.State{
				Addresses: []resolver.Address{
					{
						Addr: "test-1",
					},
				},
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			cc := &mockClientConn{}
			rs := &grpcResolver{
				r:      tc.mock(),
				target: resolver.Target{},
				cc:     cc,
			}
			rs.ResolveNow(resolver.ResolveNowOptions{})
			assert.Equal(t, tc.wantErr, cc.err)
			if cc.err != nil {
				return
			}
			state := cc.state
			assert.Equal(t, tc.wantState, state)
		})
	}
}

type mockClientConn struct {
	state resolver.State
	err   error
	resolver.ClientConn
}

func (cc *mockClientConn) UpdateState(state resolver.State) error {
	cc.state = state
	return nil
}

func (cc *mockClientConn) ReportError(err error) {
	cc.err = err
}
