package portal

import "context"

type contextKey struct {
	name string
}

var dumpDepthCtxKey = contextKey{name: "dump-depth"}

func IncrDumpDepthContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, dumpDepthCtxKey, DumpDepthFromContext(ctx)+1)
}

func DumpDepthFromContext(ctx context.Context) int {
	val := ctx.Value(dumpDepthCtxKey)
	depth, ok := val.(int)
	if !ok {
		// depth starts from 0
		return 0
	}
	return depth
}
