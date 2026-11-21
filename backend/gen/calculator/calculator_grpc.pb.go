// Code generated for calculator.proto. DO NOT EDIT.
package calculator

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type CalculatorServiceClient interface {
	Price(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error)
	Yield(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error)
	Cashflows(context.Context, *structpb.Struct, ...grpc.CallOption) (*structpb.Struct, error)
}
type calculatorServiceClient struct{ cc grpc.ClientConnInterface }

func NewCalculatorServiceClient(cc grpc.ClientConnInterface) CalculatorServiceClient {
	return &calculatorServiceClient{cc}
}
func (c *calculatorServiceClient) call(ctx context.Context, method string, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error) {
	out := new(structpb.Struct)
	err := c.cc.Invoke(ctx, "/calculator.CalculatorService/"+method, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
func (c *calculatorServiceClient) Price(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error) {
	return c.call(ctx, "Price", in, opts...)
}
func (c *calculatorServiceClient) Yield(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error) {
	return c.call(ctx, "Yield", in, opts...)
}
func (c *calculatorServiceClient) Cashflows(ctx context.Context, in *structpb.Struct, opts ...grpc.CallOption) (*structpb.Struct, error) {
	return c.call(ctx, "Cashflows", in, opts...)
}

type CalculatorServiceServer interface {
	Price(context.Context, *structpb.Struct) (*structpb.Struct, error)
	Yield(context.Context, *structpb.Struct) (*structpb.Struct, error)
	Cashflows(context.Context, *structpb.Struct) (*structpb.Struct, error)
}

func RegisterCalculatorServiceServer(s grpc.ServiceRegistrar, srv CalculatorServiceServer) {
	s.RegisterService(&CalculatorService_ServiceDesc, srv)
}
func handler(method string, invoke func(CalculatorServiceServer, context.Context, *structpb.Struct) (*structpb.Struct, error)) grpc.MethodDesc {
	return grpc.MethodDesc{MethodName: method, Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
		in := new(structpb.Struct)
		if err := dec(in); err != nil {
			return nil, err
		}
		if interceptor == nil {
			return invoke(srv.(CalculatorServiceServer), ctx, in)
		}
		info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/calculator.CalculatorService/" + method}
		return interceptor(ctx, in, info, func(ctx context.Context, req any) (any, error) {
			return invoke(srv.(CalculatorServiceServer), ctx, req.(*structpb.Struct))
		})
	}}
}

var CalculatorService_ServiceDesc = grpc.ServiceDesc{ServiceName: "calculator.CalculatorService", HandlerType: (*CalculatorServiceServer)(nil), Methods: []grpc.MethodDesc{
	handler("Price", func(s CalculatorServiceServer, c context.Context, r *structpb.Struct) (*structpb.Struct, error) {
		return s.Price(c, r)
	}),
	handler("Yield", func(s CalculatorServiceServer, c context.Context, r *structpb.Struct) (*structpb.Struct, error) {
		return s.Yield(c, r)
	}),
	handler("Cashflows", func(s CalculatorServiceServer, c context.Context, r *structpb.Struct) (*structpb.Struct, error) {
		return s.Cashflows(c, r)
	}),
}, Metadata: "calculator/calculator.proto"}
