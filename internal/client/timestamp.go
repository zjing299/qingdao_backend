package client

import (
	"context"
	"errors"

	"qingdao_backend/internal/config"
	"qingdao_backend/internal/types"
)

var (
	ErrNotConfigured   = errors.New("upstream timestamp service is not configured")
	ErrProtocolPending = errors.New("upstream timestamp protocol is pending integration")
)

// TimestampProvider separates local API behavior from the upstream wire format.
type TimestampProvider interface {
	RSA(context.Context, string) (*types.ApplyResult, error)
	SM(context.Context, string) (*types.ApplyResult, error)
	Verify(context.Context, types.VerifyRequest) (*types.VerifyResult, error)
}

type TimestampClient struct{ config config.TimestampConf }

func NewTimestampClient(c config.TimestampConf) *TimestampClient {
	return &TimestampClient{config: c}
}

// Token is an internal integration point. Implement the upstream authentication
// contract here, including token expiry and refresh, once documentation is available.
// Never return this token from the public HTTP API.
func (c *TimestampClient) Token(ctx context.Context) (string, error) {
	return "", c.pending(ctx)
}

// RSA will acquire a token internally and POST the SHA256 digest to RSAPath.
func (c *TimestampClient) RSA(ctx context.Context, rid string) (*types.ApplyResult, error) {
	return nil, c.pending(ctx)
}

// SM will acquire a token internally and POST the SM3 digest to SMPath.
func (c *TimestampClient) SM(ctx context.Context, rid string) (*types.ApplyResult, error) {
	return nil, c.pending(ctx)
}

// Verify will call VerifyPath and map the verified imprint and signature result.
// Valid must require both upstream verification and a match with RID/Algorithm.
func (c *TimestampClient) Verify(ctx context.Context, req types.VerifyRequest) (*types.VerifyResult, error) {
	return nil, c.pending(ctx)
}

func (c *TimestampClient) pending(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c.config.BaseURL == "" {
		return ErrNotConfigured
	}
	// Do not guess authentication headers, content types, or token envelopes.
	return ErrProtocolPending
}
