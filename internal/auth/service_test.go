package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeStore struct {
	mu       sync.Mutex
	users    map[string]Credentials
	names    map[int64]string
	sessions map[string]SessionRecord
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		users:    map[string]Credentials{},
		names:    map[int64]string{},
		sessions: map[string]SessionRecord{},
	}
}

func (f *fakeStore) addUser(id int64, name, passwordHash string) {
	f.users[name] = Credentials{UserID: id, PasswordHash: passwordHash}
	f.names[id] = name
}

func (f *fakeStore) UserByName(_ context.Context, name string) (Credentials, error) {
	c, ok := f.users[name]
	if !ok {
		return Credentials{}, ErrUserNotFound
	}
	return c, nil
}

func (f *fakeStore) CreateSession(_ context.Context, hash []byte, userID int64, expiresAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sessions[string(hash)] = SessionRecord{UserID: userID, UserName: f.names[userID], ExpiresAt: expiresAt}
	return nil
}

func (f *fakeStore) SessionUser(_ context.Context, hash []byte) (SessionRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rec, ok := f.sessions[string(hash)]
	if !ok {
		return SessionRecord{}, ErrSessionNotFound
	}
	return rec, nil
}

func (f *fakeStore) ExtendSession(_ context.Context, hash []byte, expiresAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	rec, ok := f.sessions[string(hash)]
	if !ok {
		return ErrSessionNotFound
	}
	rec.ExpiresAt = expiresAt
	f.sessions[string(hash)] = rec
	return nil
}

func (f *fakeStore) DeleteSession(_ context.Context, hash []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.sessions, string(hash))
	return nil
}

func TestServiceLogin(t *testing.T) {
	store := newFakeStore()
	hash, err := HashPassword("s3cret")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	store.addUser(1, "robin", hash)

	svc := NewService(store, store, true)

	tests := []struct {
		name     string
		user     string
		password string
		wantErr  error
	}{
		{name: "unknown user", user: "ghost", password: "whatever", wantErr: ErrInvalidCredentials},
		{name: "wrong password", user: "robin", password: "wrong", wantErr: ErrInvalidCredentials},
		{name: "correct credentials", user: "robin", password: "s3cret", wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie, err := svc.Login(context.Background(), tt.user, tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Login() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if cookie == nil || cookie.Value == "" {
					t.Fatalf("Login() cookie = %v, want a session cookie", cookie)
				}
				if !cookie.Secure {
					t.Errorf("cookie.Secure = false, want true")
				}
				if !cookie.HttpOnly {
					t.Errorf("cookie.HttpOnly = false, want true")
				}
			}
		})
	}
}

func TestServiceAuthenticate(t *testing.T) {
	store := newFakeStore()
	hash, err := HashPassword("s3cret")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	store.addUser(1, "robin", hash)

	svc := NewService(store, store, false)

	cookie, err := svc.Login(context.Background(), "robin", "s3cret")
	if err != nil {
		t.Fatalf("Login() error: %v", err)
	}

	user, err := svc.Authenticate(context.Background(), cookie.Value)
	if err != nil {
		t.Fatalf("Authenticate() error: %v", err)
	}
	if user.Name != "robin" || user.ID != 1 {
		t.Errorf("Authenticate() = %+v, want {ID:1 Name:robin}", user)
	}

	if _, err := svc.Authenticate(context.Background(), "not-a-real-token"); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("Authenticate() with garbage token error = %v, want %v", err, ErrSessionNotFound)
	}
}

func TestServiceAuthenticateExpiredSession(t *testing.T) {
	store := newFakeStore()
	store.addUser(1, "robin", "unused")

	svc := NewService(store, store, false)

	token, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken() error: %v", err)
	}
	if err := store.CreateSession(context.Background(), token.Hash, 1, time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}

	if _, err := svc.Authenticate(context.Background(), token.Raw); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("Authenticate() with expired session error = %v, want %v", err, ErrSessionNotFound)
	}
}

func TestServiceLogout(t *testing.T) {
	store := newFakeStore()
	hash, err := HashPassword("s3cret")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	store.addUser(1, "robin", hash)

	svc := NewService(store, store, false)

	cookie, err := svc.Login(context.Background(), "robin", "s3cret")
	if err != nil {
		t.Fatalf("Login() error: %v", err)
	}

	if err := svc.Logout(context.Background(), cookie.Value); err != nil {
		t.Fatalf("Logout() error: %v", err)
	}

	if _, err := svc.Authenticate(context.Background(), cookie.Value); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("Authenticate() after logout error = %v, want %v", err, ErrSessionNotFound)
	}
}
