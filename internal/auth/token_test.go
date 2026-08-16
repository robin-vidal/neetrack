package auth

import (
	"bytes"
	"testing"
)

func TestNewTokenHashRoundTrip(t *testing.T) {
	token, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken() error: %v", err)
	}

	hash, err := HashToken(token.Raw)
	if err != nil {
		t.Fatalf("HashToken() error: %v", err)
	}

	if !bytes.Equal(hash, token.Hash) {
		t.Errorf("HashToken(token.Raw) = %x, want %x", hash, token.Hash)
	}
}

func TestNewTokenUnique(t *testing.T) {
	a, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken() error: %v", err)
	}
	b, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken() error: %v", err)
	}

	if a.Raw == b.Raw {
		t.Errorf("NewToken() produced identical tokens across two calls")
	}
}

func TestHashTokenInvalidEncoding(t *testing.T) {
	if _, err := HashToken("not valid base64!!"); err == nil {
		t.Errorf("HashToken() error = nil, want error for malformed input")
	}
}
