package identity

import "context"

type ctxKey int

const CtxUserKey ctxKey = 1

type User struct {
	ID   uint
	Role string
}

func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, CtxUserKey, user)
}

func GetUserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(CtxUserKey).(*User)
	return user
}
