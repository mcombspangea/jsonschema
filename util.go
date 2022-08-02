package jsonschema

import (
	"context"
)

type jsonschemaCtx string

const (
	instanceCtxKey     jsonschemaCtx = "instance"
	instancePathCtxKey jsonschemaCtx = "path"
	keywordPathCtxKey  jsonschemaCtx = "keyword_path"
	scopesCtxKey       jsonschemaCtx = "scopes"
)

func GetInstance(ctx context.Context) interface{} {
	return ctx.Value(instanceCtxKey)
}

func GetScopes(ctx context.Context) []schemaRef {
	return ctx.Value(scopesCtxKey).([]schemaRef)
}

func GetKeywordPath(ctx context.Context) string {
	return ctx.Value(keywordPathCtxKey).(string)
}

func GetInstancePath(ctx context.Context) string {
	return ctx.Value(instancePathCtxKey).(string)
}

func withValues(ctx context.Context, vals ...interface{}) context.Context {
	for i := 1; i < len(vals); i += 2 {
		ctx = context.WithValue(ctx, vals[i-1], vals[i])
	}
	return ctx
}
