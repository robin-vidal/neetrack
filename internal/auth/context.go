package auth

import "context"

// User is the authenticated identity carried in a request context.
type User struct {
	ID   int64
	Name string
}

type contextKey struct{}

func withUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, contextKey{}, u)
}

// UserFromContext returns the user injected by RequireAuth, if any.
func UserFromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(contextKey{}).(User)
	return u, ok
}
