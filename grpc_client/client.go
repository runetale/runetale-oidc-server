package grpcclient

import (
	"context"
	"crypto/tls"

	"github.com/runetale/client-go/runetale/runetale/v1/login"
	"github.com/runetale/client-go/runetale/runetale/v1/oidc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	grpcMetadata "google.golang.org/grpc/metadata"
)

func NewGrpcDialOption(isTLS bool) grpc.DialOption {
	var option grpc.DialOption
	if isTLS {
		tlsCredentials := credentials.NewTLS(&tls.Config{})
		option = grpc.WithTransportCredentials(tlsCredentials)
	} else {
		option = grpc.WithTransportCredentials(insecure.NewCredentials())
	}
	return option
}

type ServerClientImpl interface {
	Login(sub, tenantID, domain, providerID, email, username, picture, token, inviteCode string) (*oidc.LoginResponse, error)
	GetInvitation(inviteCode string) (*login.GetInvitationResponse, error)
}

type ServerClient struct {
	oidcClient  oidc.OIDCServiceClient
	loginClient login.LoginServiceClient
	conn        *grpc.ClientConn
	ctx         context.Context
}

func NewServerClient(
	conn *grpc.ClientConn,
) ServerClientImpl {
	return &ServerClient{
		oidcClient:  oidc.NewOIDCServiceClient(conn),
		loginClient: login.NewLoginServiceClient(conn),
		conn:        conn,
		ctx:         context.Background(),
	}
}

func (c *ServerClient) Login(sub, tenantID, domain, providerID, email, username, picture, token, inviteCode string) (*oidc.LoginResponse, error) {
	ctx := grpcMetadata.AppendToOutgoingContext(c.ctx, "authorization", "Bearer "+token)
	return c.oidcClient.Login(ctx, &oidc.LoginRequest{Sub: sub, TenantID: tenantID, Doamin: domain, ProviderID: providerID, Email: email, Username: username, Picture: picture, InviteCode: inviteCode})
}

func (c *ServerClient) GetInvitation(inviteCode string) (*login.GetInvitationResponse, error) {
	return c.loginClient.GetInvitation(c.ctx, &login.GetInvitationRequest{InviteCode: inviteCode})
}
