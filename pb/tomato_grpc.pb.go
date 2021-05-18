// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TomatoServiceClient is the client API for TomatoService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TomatoServiceClient interface {
	Start(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*timestamppb.Timestamp, error)
	Stop(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*durationpb.Duration, error)
	Remaining(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*durationpb.Duration, error)
	Running(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wrapperspb.BoolValue, error)
}

type tomatoServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTomatoServiceClient(cc grpc.ClientConnInterface) TomatoServiceClient {
	return &tomatoServiceClient{cc}
}

func (c *tomatoServiceClient) Start(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*timestamppb.Timestamp, error) {
	out := new(timestamppb.Timestamp)
	err := c.cc.Invoke(ctx, "/tomato.pb.TomatoService/Start", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tomatoServiceClient) Stop(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*durationpb.Duration, error) {
	out := new(durationpb.Duration)
	err := c.cc.Invoke(ctx, "/tomato.pb.TomatoService/Stop", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tomatoServiceClient) Remaining(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*durationpb.Duration, error) {
	out := new(durationpb.Duration)
	err := c.cc.Invoke(ctx, "/tomato.pb.TomatoService/Remaining", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tomatoServiceClient) Running(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wrapperspb.BoolValue, error) {
	out := new(wrapperspb.BoolValue)
	err := c.cc.Invoke(ctx, "/tomato.pb.TomatoService/Running", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TomatoServiceServer is the server API for TomatoService service.
// All implementations must embed UnimplementedTomatoServiceServer
// for forward compatibility
type TomatoServiceServer interface {
	Start(context.Context, *emptypb.Empty) (*timestamppb.Timestamp, error)
	Stop(context.Context, *emptypb.Empty) (*durationpb.Duration, error)
	Remaining(context.Context, *emptypb.Empty) (*durationpb.Duration, error)
	Running(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error)
	mustEmbedUnimplementedTomatoServiceServer()
}

// UnimplementedTomatoServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTomatoServiceServer struct {
}

func (UnimplementedTomatoServiceServer) Start(context.Context, *emptypb.Empty) (*timestamppb.Timestamp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Start not implemented")
}
func (UnimplementedTomatoServiceServer) Stop(context.Context, *emptypb.Empty) (*durationpb.Duration, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stop not implemented")
}
func (UnimplementedTomatoServiceServer) Remaining(context.Context, *emptypb.Empty) (*durationpb.Duration, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Remaining not implemented")
}
func (UnimplementedTomatoServiceServer) Running(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Running not implemented")
}
func (UnimplementedTomatoServiceServer) mustEmbedUnimplementedTomatoServiceServer() {}

// UnsafeTomatoServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TomatoServiceServer will
// result in compilation errors.
type UnsafeTomatoServiceServer interface {
	mustEmbedUnimplementedTomatoServiceServer()
}

func RegisterTomatoServiceServer(s grpc.ServiceRegistrar, srv TomatoServiceServer) {
	s.RegisterService(&TomatoService_ServiceDesc, srv)
}

func _TomatoService_Start_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TomatoServiceServer).Start(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tomato.pb.TomatoService/Start",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TomatoServiceServer).Start(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _TomatoService_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TomatoServiceServer).Stop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tomato.pb.TomatoService/Stop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TomatoServiceServer).Stop(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _TomatoService_Remaining_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TomatoServiceServer).Remaining(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tomato.pb.TomatoService/Remaining",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TomatoServiceServer).Remaining(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _TomatoService_Running_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TomatoServiceServer).Running(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tomato.pb.TomatoService/Running",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TomatoServiceServer).Running(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// TomatoService_ServiceDesc is the grpc.ServiceDesc for TomatoService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TomatoService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tomato.pb.TomatoService",
	HandlerType: (*TomatoServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Start",
			Handler:    _TomatoService_Start_Handler,
		},
		{
			MethodName: "Stop",
			Handler:    _TomatoService_Stop_Handler,
		},
		{
			MethodName: "Remaining",
			Handler:    _TomatoService_Remaining_Handler,
		},
		{
			MethodName: "Running",
			Handler:    _TomatoService_Running_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tomato.proto",
}