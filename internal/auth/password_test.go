package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{name: "correct password", password: "correct horse battery staple", want: true},
		{name: "wrong password", password: "wrong password", want: false},
		{name: "empty password", password: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := VerifyPassword(hash, tt.password)
			if err != nil {
				t.Fatalf("VerifyPassword() error: %v", err)
			}
			if ok != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", ok, tt.want)
			}
		})
	}
}

func TestHashPasswordUniqueSalt(t *testing.T) {
	a, err := HashPassword("same password")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	b, err := HashPassword("same password")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	if a == b {
		t.Errorf("HashPassword() produced identical output for two calls, want distinct salts")
	}
}

func TestVerifyPasswordMalformedHash(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{name: "empty", hash: ""},
		{name: "not phc", hash: "not-a-valid-hash"},
		{name: "wrong algorithm", hash: "$argon2i$v=19$m=65536,t=1,p=4$c2FsdA$aGFzaA"},
		{name: "bad params", hash: "$argon2id$v=19$m=notanumber,t=1,p=4$c2FsdA$aGFzaA"},
		{name: "bad base64 salt", hash: "$argon2id$v=19$m=65536,t=1,p=4$not-base64!$aGFzaA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := VerifyPassword(tt.hash, "anything"); err == nil {
				t.Errorf("VerifyPassword() error = nil, want error")
			}
		})
	}
}
