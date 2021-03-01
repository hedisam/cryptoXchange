// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package grpcapi

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// PriceServiceClient is the client API for PriceService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PriceServiceClient interface {
	StreamPrice(ctx context.Context, in *QuoteRequest, opts ...grpc.CallOption) (PriceService_StreamPriceClient, error)
}

type priceServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPriceServiceClient(cc grpc.ClientConnInterface) PriceServiceClient {
	return &priceServiceClient{cc}
}

func (c *priceServiceClient) StreamPrice(ctx context.Context, in *QuoteRequest, opts ...grpc.CallOption) (PriceService_StreamPriceClient, error) {
	stream, err := c.cc.NewStream(ctx, &PriceService_ServiceDesc.Streams[0], "/grpcapi.PriceService/StreamPrice", opts...)
	if err != nil {
		return nil, err
	}
	x := &priceServiceStreamPriceClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type PriceService_StreamPriceClient interface {
	Recv() (*QuoteReply, error)
	grpc.ClientStream
}

type priceServiceStreamPriceClient struct {
	grpc.ClientStream
}

func (x *priceServiceStreamPriceClient) Recv() (*QuoteReply, error) {
	m := new(QuoteReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// PriceServiceServer is the server API for PriceService service.
// All implementations must embed UnimplementedPriceServiceServer
// for forward compatibility
type PriceServiceServer interface {
	StreamPrice(*QuoteRequest, PriceService_StreamPriceServer) error
	mustEmbedUnimplementedPriceServiceServer()
}

// UnimplementedPriceServiceServer must be embedded to have forward compatible implementations.
type UnimplementedPriceServiceServer struct {
}

func (UnimplementedPriceServiceServer) StreamPrice(*QuoteRequest, PriceService_StreamPriceServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamPrice not implemented")
}
func (UnimplementedPriceServiceServer) mustEmbedUnimplementedPriceServiceServer() {}

// UnsafePriceServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PriceServiceServer will
// result in compilation errors.
type UnsafePriceServiceServer interface {
	mustEmbedUnimplementedPriceServiceServer()
}

func RegisterPriceServiceServer(s grpc.ServiceRegistrar, srv PriceServiceServer) {
	s.RegisterService(&PriceService_ServiceDesc, srv)
}

func _PriceService_StreamPrice_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(QuoteRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PriceServiceServer).StreamPrice(m, &priceServiceStreamPriceServer{stream})
}

type PriceService_StreamPriceServer interface {
	Send(*QuoteReply) error
	grpc.ServerStream
}

type priceServiceStreamPriceServer struct {
	grpc.ServerStream
}

func (x *priceServiceStreamPriceServer) Send(m *QuoteReply) error {
	return x.ServerStream.SendMsg(m)
}

// PriceService_ServiceDesc is the grpc.ServiceDesc for PriceService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PriceService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpcapi.PriceService",
	HandlerType: (*PriceServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamPrice",
			Handler:       _PriceService_StreamPrice_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "userservice/grpcapi/grpcapi.proto",
}
