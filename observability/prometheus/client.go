package prometheus

import (
	"context"
	"google.golang.org/grpc"
)

type ClientInterceptorBuilder struct {
	NameSpace string
	Subsystem string
	Name      string
	Help      string
}

func (c ClientInterceptorBuilder) BuildUnary() grpc.UnaryClientInterceptor {
	//summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
	//	Namespace: c.NameSpace,
	//	Subsystem: c.Subsystem,
	//	Name:      c.Name,
	//	Help:      c.Help,
	//	ConstLabels: map[string]string{
	//		"component": "client",
	//		"address":   "",
	//	},
	//}, []string{})
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return nil
	}

}
