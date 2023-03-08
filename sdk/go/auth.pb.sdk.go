// Code generated by protoc-gen-go-flipt-sdk. DO NOT EDIT.

package sdk

import (
	context "context"
	auth "go.flipt.io/flipt/rpc/flipt/auth"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type AuthTransport interface {
	ListAuthenticationMethods(context.Context, *emptypb.Empty) (*auth.ListAuthenticationMethodsResponse, error)
	GetAuthenticationSelf(context.Context, *emptypb.Empty) (*auth.Authentication, error)
	GetAuthentication(context.Context, *auth.GetAuthenticationRequest) (*auth.Authentication, error)
	ListAuthentications(context.Context, *auth.ListAuthenticationsRequest) (*auth.ListAuthenticationsResponse, error)
	DeleteAuthentication(context.Context, *auth.DeleteAuthenticationRequest) (*emptypb.Empty, error)
	ExpireAuthenticationSelf(context.Context, *auth.ExpireAuthenticationSelfRequest) (*emptypb.Empty, error)
	CreateToken(context.Context, *auth.CreateTokenRequest) (*auth.CreateTokenResponse, error)
	AuthorizeURL(context.Context, *auth.AuthorizeURLRequest) (*auth.AuthorizeURLResponse, error)
	Callback(context.Context, *auth.CallbackRequest) (*auth.CallbackResponse, error)
	VerifyServiceAccount(context.Context, *auth.VerifyServiceAccountRequest) (*auth.VerifyServiceAccountResponse, error)
}

type Auth struct {
	transport AuthTransport
}

func (x *Auth) ListAuthenticationMethods(ctx context.Context) (*auth.ListAuthenticationMethodsResponse, error) {
	return x.transport.ListAuthenticationMethods(ctx, &emptypb.Empty{})
}

func (x *Auth) GetAuthenticationSelf(ctx context.Context) (*auth.Authentication, error) {
	return x.transport.GetAuthenticationSelf(ctx, &emptypb.Empty{})
}

func (x *Auth) GetAuthentication(ctx context.Context, id string) (*auth.Authentication, error) {
	return x.transport.GetAuthentication(ctx, &auth.GetAuthenticationRequest{Id: id})
}

func (x *Auth) ListAuthentications(ctx context.Context, v *auth.ListAuthenticationsRequest) (*auth.ListAuthenticationsResponse, error) {
	return x.transport.ListAuthentications(ctx, v)
}

func (x *Auth) DeleteAuthentication(ctx context.Context, id string) error {
	_, err := x.transport.DeleteAuthentication(ctx, &auth.DeleteAuthenticationRequest{Id: id})
	return err
}

func (x *Auth) ExpireAuthenticationSelf(ctx context.Context, expiresAt *timestamppb.Timestamp) error {
	_, err := x.transport.ExpireAuthenticationSelf(ctx, &auth.ExpireAuthenticationSelfRequest{ExpiresAt: expiresAt})
	return err
}

func (x *Auth) CreateToken(ctx context.Context, v *auth.CreateTokenRequest) (*auth.CreateTokenResponse, error) {
	return x.transport.CreateToken(ctx, v)
}

func (x *Auth) AuthorizeURL(ctx context.Context, v *auth.AuthorizeURLRequest) (*auth.AuthorizeURLResponse, error) {
	return x.transport.AuthorizeURL(ctx, v)
}

func (x *Auth) Callback(ctx context.Context, v *auth.CallbackRequest) (*auth.CallbackResponse, error) {
	return x.transport.Callback(ctx, v)
}

func (x *Auth) VerifyServiceAccount(ctx context.Context, serviceAccountToken string) (*auth.VerifyServiceAccountResponse, error) {
	return x.transport.VerifyServiceAccount(ctx, &auth.VerifyServiceAccountRequest{ServiceAccountToken: serviceAccountToken})
}
