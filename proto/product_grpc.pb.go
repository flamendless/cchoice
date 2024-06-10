// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v5.27.0
// source: proto/product.proto

package pb

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

// ProductServiceClient is the client API for ProductService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ProductServiceClient interface {
	Void(ctx context.Context, in *VoidParam, opts ...grpc.CallOption) (*VoidReturn, error)
	GetProductCategoryByID(ctx context.Context, in *ID, opts ...grpc.CallOption) (*ProductCategory, error)
	GetProductSpecsByID(ctx context.Context, in *ID, opts ...grpc.CallOption) (*ProductSpecs, error)
	GetProductByID(ctx context.Context, in *ID, opts ...grpc.CallOption) (*Product, error)
}

type productServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewProductServiceClient(cc grpc.ClientConnInterface) ProductServiceClient {
	return &productServiceClient{cc}
}

func (c *productServiceClient) Void(ctx context.Context, in *VoidParam, opts ...grpc.CallOption) (*VoidReturn, error) {
	out := new(VoidReturn)
	err := c.cc.Invoke(ctx, "/proto.ProductService/Void", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *productServiceClient) GetProductCategoryByID(ctx context.Context, in *ID, opts ...grpc.CallOption) (*ProductCategory, error) {
	out := new(ProductCategory)
	err := c.cc.Invoke(ctx, "/proto.ProductService/GetProductCategoryByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *productServiceClient) GetProductSpecsByID(ctx context.Context, in *ID, opts ...grpc.CallOption) (*ProductSpecs, error) {
	out := new(ProductSpecs)
	err := c.cc.Invoke(ctx, "/proto.ProductService/GetProductSpecsByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *productServiceClient) GetProductByID(ctx context.Context, in *ID, opts ...grpc.CallOption) (*Product, error) {
	out := new(Product)
	err := c.cc.Invoke(ctx, "/proto.ProductService/GetProductByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProductServiceServer is the server API for ProductService service.
// All implementations must embed UnimplementedProductServiceServer
// for forward compatibility
type ProductServiceServer interface {
	Void(context.Context, *VoidParam) (*VoidReturn, error)
	GetProductCategoryByID(context.Context, *ID) (*ProductCategory, error)
	GetProductSpecsByID(context.Context, *ID) (*ProductSpecs, error)
	GetProductByID(context.Context, *ID) (*Product, error)
	mustEmbedUnimplementedProductServiceServer()
}

// UnimplementedProductServiceServer must be embedded to have forward compatible implementations.
type UnimplementedProductServiceServer struct {
}

func (UnimplementedProductServiceServer) Void(context.Context, *VoidParam) (*VoidReturn, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Void not implemented")
}
func (UnimplementedProductServiceServer) GetProductCategoryByID(context.Context, *ID) (*ProductCategory, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProductCategoryByID not implemented")
}
func (UnimplementedProductServiceServer) GetProductSpecsByID(context.Context, *ID) (*ProductSpecs, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProductSpecsByID not implemented")
}
func (UnimplementedProductServiceServer) GetProductByID(context.Context, *ID) (*Product, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProductByID not implemented")
}
func (UnimplementedProductServiceServer) mustEmbedUnimplementedProductServiceServer() {}

// UnsafeProductServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ProductServiceServer will
// result in compilation errors.
type UnsafeProductServiceServer interface {
	mustEmbedUnimplementedProductServiceServer()
}

func RegisterProductServiceServer(s grpc.ServiceRegistrar, srv ProductServiceServer) {
	s.RegisterService(&ProductService_ServiceDesc, srv)
}

func _ProductService_Void_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VoidParam)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServiceServer).Void(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ProductService/Void",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductServiceServer).Void(ctx, req.(*VoidParam))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductService_GetProductCategoryByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServiceServer).GetProductCategoryByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ProductService/GetProductCategoryByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductServiceServer).GetProductCategoryByID(ctx, req.(*ID))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductService_GetProductSpecsByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServiceServer).GetProductSpecsByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ProductService/GetProductSpecsByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductServiceServer).GetProductSpecsByID(ctx, req.(*ID))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductService_GetProductByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServiceServer).GetProductByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ProductService/GetProductByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductServiceServer).GetProductByID(ctx, req.(*ID))
	}
	return interceptor(ctx, in, info, handler)
}

// ProductService_ServiceDesc is the grpc.ServiceDesc for ProductService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ProductService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ProductService",
	HandlerType: (*ProductServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Void",
			Handler:    _ProductService_Void_Handler,
		},
		{
			MethodName: "GetProductCategoryByID",
			Handler:    _ProductService_GetProductCategoryByID_Handler,
		},
		{
			MethodName: "GetProductSpecsByID",
			Handler:    _ProductService_GetProductSpecsByID_Handler,
		},
		{
			MethodName: "GetProductByID",
			Handler:    _ProductService_GetProductByID_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/product.proto",
}