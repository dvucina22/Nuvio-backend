package host

import "context"

type ctxKey string

const authHeaderKey ctxKey = "auth_header"

func WithAuthHeader(ctx context.Context, header string) context.Context {
	return context.WithValue(ctx, authHeaderKey, header)
}

func AuthHeaderFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(authHeaderKey)
	s, ok := v.(string)
	return s, ok && s != ""
}
