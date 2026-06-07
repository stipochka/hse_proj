package jwks

import (
	"context"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Parser is the interface handlers use to validate JWTs.
type Parser interface {
	Parse(tokenStr string) (jwt.MapClaims, error)
}

// Client wraps the JWKS keyfunc and exposes a single Parse method.
type Client struct {
	kf keyfunc.Keyfunc
}

// New creates a client that automatically refreshes public keys from the JWKS URL.
func New(ctx context.Context, jwksURL string) (*Client, error) {
	kf, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("jwks init: %w", err)
	}
	return &Client{kf: kf}, nil
}

// Parse validates the JWT token string and returns its claims.
func (c *Client) Parse(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, c.kf.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}
	return claims, nil
}
