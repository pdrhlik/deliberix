package identity

import "context"

type Actor struct {
	UserID        *uint
	AnonSessionID *string
	User          *User
}

func (a *Actor) IsAnon() bool          { return a != nil && a.AnonSessionID != nil }
func (a *Actor) IsAuthenticated() bool { return a != nil && a.UserID != nil }

type actorKey struct{}

func ContextWithActor(ctx context.Context, a *Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, a)
}

func GetActorFromContext(ctx context.Context) *Actor {
	a, _ := ctx.Value(actorKey{}).(*Actor)
	return a
}
