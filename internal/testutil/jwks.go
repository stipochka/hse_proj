// Package testutil provides helpers for integration tests.
package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWKSServer is a test JWKS endpoint backed by a generated RSA key pair.
// Use Sign to create valid tokens that the real jwks.Client will accept.
type JWKSServer struct {
	Server     *httptest.Server
	privateKey *rsa.PrivateKey
}

// NewJWKSServer generates an RSA-2048 key pair and starts a test HTTP server
// that serves the corresponding JWKS document at "/".
func NewJWKSServer() (*JWKSServer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	n := base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(bigEndianUint(key.PublicKey.E))

	jwksBytes, _ := json.Marshal(map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"use": "sig",
			"kid": "test-key",
			"alg": "RS256",
			"n":   n,
			"e":   e,
		}},
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(jwksBytes)
	}))

	return &JWKSServer{Server: srv, privateKey: key}, nil
}

// Sign returns a signed JWT mirroring what Keycloak issues:
// role in realm_access.roles and group membership in the `groups` claim.
// role must be one of "student", "group_admin", "super_admin".
func (j *JWKSServer) Sign(sub, email, role string, groups ...string) string {
	claims := jwt.MapClaims{
		"sub":   sub,
		"email": email,
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"realm_access": map[string]any{
			"roles": []any{role},
		},
	}
	if len(groups) > 0 {
		g := make([]any, len(groups))
		for i, v := range groups {
			g[i] = v
		}
		claims["groups"] = g
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = "test-key"
	signed, _ := tok.SignedString(j.privateKey)
	return signed
}

// bigEndianUint encodes an int as a minimal big-endian byte slice (no leading zeros).
func bigEndianUint(v int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	for len(b) > 1 && b[0] == 0 {
		b = b[1:]
	}
	return b
}
