// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: proto/user_dependency.proto

package usergrpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	UserService_Register_FullMethodName                     = "/user.UserService/Register"
	UserService_Login_FullMethodName                        = "/user.UserService/Login"
	UserService_Logout_FullMethodName                       = "/user.UserService/Logout"
	UserService_GetProfile_FullMethodName                   = "/user.UserService/GetProfile"
	UserService_UpdateProfile_FullMethodName                = "/user.UserService/UpdateProfile"
	UserService_ChangePassword_FullMethodName               = "/user.UserService/ChangePassword"
	UserService_DeleteUser_FullMethodName                   = "/user.UserService/DeleteUser"
	UserService_DeactivateUser_FullMethodName               = "/user.UserService/DeactivateUser"
	UserService_RequestEmailVerification_FullMethodName     = "/user.UserService/RequestEmailVerification"
	UserService_VerifyEmail_FullMethodName                  = "/user.UserService/VerifyEmail"
	UserService_CheckEmailVerificationStatus_FullMethodName = "/user.UserService/CheckEmailVerificationStatus"
	UserService_AdminDeleteUser_FullMethodName              = "/user.UserService/AdminDeleteUser"
	UserService_AdminListUsers_FullMethodName               = "/user.UserService/AdminListUsers"
	UserService_AdminSearchUsers_FullMethodName             = "/user.UserService/AdminSearchUsers"
	UserService_AdminUpdateUserRole_FullMethodName          = "/user.UserService/AdminUpdateUserRole"
	UserService_AdminSetUserActiveStatus_FullMethodName     = "/user.UserService/AdminSetUserActiveStatus"
)

// UserServiceClient is the client API for UserService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserServiceClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
	Logout(ctx context.Context, in *LogoutRequest, opts ...grpc.CallOption) (*LogoutResponse, error)
	GetProfile(ctx context.Context, in *GetProfileRequest, opts ...grpc.CallOption) (*GetProfileResponse, error)
	UpdateProfile(ctx context.Context, in *UpdateProfileRequest, opts ...grpc.CallOption) (*UpdateProfileResponse, error)
	ChangePassword(ctx context.Context, in *ChangePasswordRequest, opts ...grpc.CallOption) (*ChangePasswordResponse, error)
	DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserResponse, error)
	DeactivateUser(ctx context.Context, in *DeactivateUserRequest, opts ...grpc.CallOption) (*DeactivateUserResponse, error)
	// Email Verification RPCs
	RequestEmailVerification(ctx context.Context, in *RequestEmailVerificationRequest, opts ...grpc.CallOption) (*RequestEmailVerificationResponse, error)
	VerifyEmail(ctx context.Context, in *VerifyEmailRequest, opts ...grpc.CallOption) (*VerifyEmailResponse, error)
	CheckEmailVerificationStatus(ctx context.Context, in *CheckEmailVerificationStatusRequest, opts ...grpc.CallOption) (*CheckEmailVerificationStatusResponse, error)
	// Admin methods
	AdminDeleteUser(ctx context.Context, in *AdminDeleteUserRequest, opts ...grpc.CallOption) (*AdminDeleteUserResponse, error)
	AdminListUsers(ctx context.Context, in *AdminListUsersRequest, opts ...grpc.CallOption) (*AdminListUsersResponse, error)
	AdminSearchUsers(ctx context.Context, in *AdminSearchUsersRequest, opts ...grpc.CallOption) (*AdminSearchUsersResponse, error)
	AdminUpdateUserRole(ctx context.Context, in *AdminUpdateUserRoleRequest, opts ...grpc.CallOption) (*AdminUpdateUserRoleResponse, error)
	AdminSetUserActiveStatus(ctx context.Context, in *AdminSetUserActiveStatusRequest, opts ...grpc.CallOption) (*AdminSetUserActiveStatusResponse, error)
}

type userServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserServiceClient(cc grpc.ClientConnInterface) UserServiceClient {
	return &userServiceClient{cc}
}

func (c *userServiceClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, UserService_Register_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, UserService_Login_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) Logout(ctx context.Context, in *LogoutRequest, opts ...grpc.CallOption) (*LogoutResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LogoutResponse)
	err := c.cc.Invoke(ctx, UserService_Logout_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) GetProfile(ctx context.Context, in *GetProfileRequest, opts ...grpc.CallOption) (*GetProfileResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetProfileResponse)
	err := c.cc.Invoke(ctx, UserService_GetProfile_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) UpdateProfile(ctx context.Context, in *UpdateProfileRequest, opts ...grpc.CallOption) (*UpdateProfileResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateProfileResponse)
	err := c.cc.Invoke(ctx, UserService_UpdateProfile_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) ChangePassword(ctx context.Context, in *ChangePasswordRequest, opts ...grpc.CallOption) (*ChangePasswordResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ChangePasswordResponse)
	err := c.cc.Invoke(ctx, UserService_ChangePassword_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) DeleteUser(ctx context.Context, in *DeleteUserRequest, opts ...grpc.CallOption) (*DeleteUserResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteUserResponse)
	err := c.cc.Invoke(ctx, UserService_DeleteUser_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) DeactivateUser(ctx context.Context, in *DeactivateUserRequest, opts ...grpc.CallOption) (*DeactivateUserResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeactivateUserResponse)
	err := c.cc.Invoke(ctx, UserService_DeactivateUser_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) RequestEmailVerification(ctx context.Context, in *RequestEmailVerificationRequest, opts ...grpc.CallOption) (*RequestEmailVerificationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RequestEmailVerificationResponse)
	err := c.cc.Invoke(ctx, UserService_RequestEmailVerification_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) VerifyEmail(ctx context.Context, in *VerifyEmailRequest, opts ...grpc.CallOption) (*VerifyEmailResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(VerifyEmailResponse)
	err := c.cc.Invoke(ctx, UserService_VerifyEmail_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) CheckEmailVerificationStatus(ctx context.Context, in *CheckEmailVerificationStatusRequest, opts ...grpc.CallOption) (*CheckEmailVerificationStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CheckEmailVerificationStatusResponse)
	err := c.cc.Invoke(ctx, UserService_CheckEmailVerificationStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) AdminDeleteUser(ctx context.Context, in *AdminDeleteUserRequest, opts ...grpc.CallOption) (*AdminDeleteUserResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AdminDeleteUserResponse)
	err := c.cc.Invoke(ctx, UserService_AdminDeleteUser_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) AdminListUsers(ctx context.Context, in *AdminListUsersRequest, opts ...grpc.CallOption) (*AdminListUsersResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AdminListUsersResponse)
	err := c.cc.Invoke(ctx, UserService_AdminListUsers_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) AdminSearchUsers(ctx context.Context, in *AdminSearchUsersRequest, opts ...grpc.CallOption) (*AdminSearchUsersResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AdminSearchUsersResponse)
	err := c.cc.Invoke(ctx, UserService_AdminSearchUsers_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) AdminUpdateUserRole(ctx context.Context, in *AdminUpdateUserRoleRequest, opts ...grpc.CallOption) (*AdminUpdateUserRoleResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AdminUpdateUserRoleResponse)
	err := c.cc.Invoke(ctx, UserService_AdminUpdateUserRole_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) AdminSetUserActiveStatus(ctx context.Context, in *AdminSetUserActiveStatusRequest, opts ...grpc.CallOption) (*AdminSetUserActiveStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AdminSetUserActiveStatusResponse)
	err := c.cc.Invoke(ctx, UserService_AdminSetUserActiveStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserServiceServer is the server API for UserService service.
// All implementations must embed UnimplementedUserServiceServer
// for forward compatibility.
type UserServiceServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
	Logout(context.Context, *LogoutRequest) (*LogoutResponse, error)
	GetProfile(context.Context, *GetProfileRequest) (*GetProfileResponse, error)
	UpdateProfile(context.Context, *UpdateProfileRequest) (*UpdateProfileResponse, error)
	ChangePassword(context.Context, *ChangePasswordRequest) (*ChangePasswordResponse, error)
	DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserResponse, error)
	DeactivateUser(context.Context, *DeactivateUserRequest) (*DeactivateUserResponse, error)
	// Email Verification RPCs
	RequestEmailVerification(context.Context, *RequestEmailVerificationRequest) (*RequestEmailVerificationResponse, error)
	VerifyEmail(context.Context, *VerifyEmailRequest) (*VerifyEmailResponse, error)
	CheckEmailVerificationStatus(context.Context, *CheckEmailVerificationStatusRequest) (*CheckEmailVerificationStatusResponse, error)
	// Admin methods
	AdminDeleteUser(context.Context, *AdminDeleteUserRequest) (*AdminDeleteUserResponse, error)
	AdminListUsers(context.Context, *AdminListUsersRequest) (*AdminListUsersResponse, error)
	AdminSearchUsers(context.Context, *AdminSearchUsersRequest) (*AdminSearchUsersResponse, error)
	AdminUpdateUserRole(context.Context, *AdminUpdateUserRoleRequest) (*AdminUpdateUserRoleResponse, error)
	AdminSetUserActiveStatus(context.Context, *AdminSetUserActiveStatusRequest) (*AdminSetUserActiveStatusResponse, error)
	mustEmbedUnimplementedUserServiceServer()
}

// UnimplementedUserServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedUserServiceServer struct{}

func (UnimplementedUserServiceServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedUserServiceServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedUserServiceServer) Logout(context.Context, *LogoutRequest) (*LogoutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logout not implemented")
}
func (UnimplementedUserServiceServer) GetProfile(context.Context, *GetProfileRequest) (*GetProfileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProfile not implemented")
}
func (UnimplementedUserServiceServer) UpdateProfile(context.Context, *UpdateProfileRequest) (*UpdateProfileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProfile not implemented")
}
func (UnimplementedUserServiceServer) ChangePassword(context.Context, *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangePassword not implemented")
}
func (UnimplementedUserServiceServer) DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}
func (UnimplementedUserServiceServer) DeactivateUser(context.Context, *DeactivateUserRequest) (*DeactivateUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeactivateUser not implemented")
}
func (UnimplementedUserServiceServer) RequestEmailVerification(context.Context, *RequestEmailVerificationRequest) (*RequestEmailVerificationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestEmailVerification not implemented")
}
func (UnimplementedUserServiceServer) VerifyEmail(context.Context, *VerifyEmailRequest) (*VerifyEmailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyEmail not implemented")
}
func (UnimplementedUserServiceServer) CheckEmailVerificationStatus(context.Context, *CheckEmailVerificationStatusRequest) (*CheckEmailVerificationStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckEmailVerificationStatus not implemented")
}
func (UnimplementedUserServiceServer) AdminDeleteUser(context.Context, *AdminDeleteUserRequest) (*AdminDeleteUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AdminDeleteUser not implemented")
}
func (UnimplementedUserServiceServer) AdminListUsers(context.Context, *AdminListUsersRequest) (*AdminListUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AdminListUsers not implemented")
}
func (UnimplementedUserServiceServer) AdminSearchUsers(context.Context, *AdminSearchUsersRequest) (*AdminSearchUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AdminSearchUsers not implemented")
}
func (UnimplementedUserServiceServer) AdminUpdateUserRole(context.Context, *AdminUpdateUserRoleRequest) (*AdminUpdateUserRoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AdminUpdateUserRole not implemented")
}
func (UnimplementedUserServiceServer) AdminSetUserActiveStatus(context.Context, *AdminSetUserActiveStatusRequest) (*AdminSetUserActiveStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AdminSetUserActiveStatus not implemented")
}
func (UnimplementedUserServiceServer) mustEmbedUnimplementedUserServiceServer() {}
func (UnimplementedUserServiceServer) testEmbeddedByValue()                     {}

// UnsafeUserServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserServiceServer will
// result in compilation errors.
type UnsafeUserServiceServer interface {
	mustEmbedUnimplementedUserServiceServer()
}

func RegisterUserServiceServer(s grpc.ServiceRegistrar, srv UserServiceServer) {
	// If the following call pancis, it indicates UnimplementedUserServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&UserService_ServiceDesc, srv)
}

func _UserService_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_Register_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_Login_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_Logout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogoutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Logout(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_Logout_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Logout(ctx, req.(*LogoutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_GetProfile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetProfileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).GetProfile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_GetProfile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).GetProfile(ctx, req.(*GetProfileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_UpdateProfile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateProfileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).UpdateProfile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_UpdateProfile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).UpdateProfile(ctx, req.(*UpdateProfileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_ChangePassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangePasswordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).ChangePassword(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_ChangePassword_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).ChangePassword(ctx, req.(*ChangePasswordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_DeleteUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).DeleteUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_DeleteUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).DeleteUser(ctx, req.(*DeleteUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_DeactivateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeactivateUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).DeactivateUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_DeactivateUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).DeactivateUser(ctx, req.(*DeactivateUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_RequestEmailVerification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestEmailVerificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).RequestEmailVerification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_RequestEmailVerification_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).RequestEmailVerification(ctx, req.(*RequestEmailVerificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_VerifyEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyEmailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).VerifyEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_VerifyEmail_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).VerifyEmail(ctx, req.(*VerifyEmailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_CheckEmailVerificationStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckEmailVerificationStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).CheckEmailVerificationStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_CheckEmailVerificationStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).CheckEmailVerificationStatus(ctx, req.(*CheckEmailVerificationStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_AdminDeleteUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AdminDeleteUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).AdminDeleteUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_AdminDeleteUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).AdminDeleteUser(ctx, req.(*AdminDeleteUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_AdminListUsers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AdminListUsersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).AdminListUsers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_AdminListUsers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).AdminListUsers(ctx, req.(*AdminListUsersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_AdminSearchUsers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AdminSearchUsersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).AdminSearchUsers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_AdminSearchUsers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).AdminSearchUsers(ctx, req.(*AdminSearchUsersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_AdminUpdateUserRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AdminUpdateUserRoleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).AdminUpdateUserRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_AdminUpdateUserRole_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).AdminUpdateUserRole(ctx, req.(*AdminUpdateUserRoleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_AdminSetUserActiveStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AdminSetUserActiveStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).AdminSetUserActiveStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_AdminSetUserActiveStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).AdminSetUserActiveStatus(ctx, req.(*AdminSetUserActiveStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// UserService_ServiceDesc is the grpc.ServiceDesc for UserService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "user.UserService",
	HandlerType: (*UserServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _UserService_Register_Handler,
		},
		{
			MethodName: "Login",
			Handler:    _UserService_Login_Handler,
		},
		{
			MethodName: "Logout",
			Handler:    _UserService_Logout_Handler,
		},
		{
			MethodName: "GetProfile",
			Handler:    _UserService_GetProfile_Handler,
		},
		{
			MethodName: "UpdateProfile",
			Handler:    _UserService_UpdateProfile_Handler,
		},
		{
			MethodName: "ChangePassword",
			Handler:    _UserService_ChangePassword_Handler,
		},
		{
			MethodName: "DeleteUser",
			Handler:    _UserService_DeleteUser_Handler,
		},
		{
			MethodName: "DeactivateUser",
			Handler:    _UserService_DeactivateUser_Handler,
		},
		{
			MethodName: "RequestEmailVerification",
			Handler:    _UserService_RequestEmailVerification_Handler,
		},
		{
			MethodName: "VerifyEmail",
			Handler:    _UserService_VerifyEmail_Handler,
		},
		{
			MethodName: "CheckEmailVerificationStatus",
			Handler:    _UserService_CheckEmailVerificationStatus_Handler,
		},
		{
			MethodName: "AdminDeleteUser",
			Handler:    _UserService_AdminDeleteUser_Handler,
		},
		{
			MethodName: "AdminListUsers",
			Handler:    _UserService_AdminListUsers_Handler,
		},
		{
			MethodName: "AdminSearchUsers",
			Handler:    _UserService_AdminSearchUsers_Handler,
		},
		{
			MethodName: "AdminUpdateUserRole",
			Handler:    _UserService_AdminUpdateUserRole_Handler,
		},
		{
			MethodName: "AdminSetUserActiveStatus",
			Handler:    _UserService_AdminSetUserActiveStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/user_dependency.proto",
}
