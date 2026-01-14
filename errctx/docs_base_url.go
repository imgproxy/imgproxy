package errctx

import "context"

// DocsBaseURLContextKey represents context.Context key to set documentation URL
// for current handler
type DocsBaseURLContextKey struct{}

// DocsBaseURL returns documentation base url set for context or default
func DocsBaseURL(ctx context.Context, def string) string {
	if url, ok := ctx.Value(DocsBaseURLContextKey{}).(string); ok && len(url) > 0 {
		return url
	}

	return def
}
