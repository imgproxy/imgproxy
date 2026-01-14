package errctx

type Option func(*ErrorContext)

func WithPrefix(prefix string) Option {
	return func(ec *ErrorContext) {
		if len(ec.prefix) > 0 {
			ec.prefix = prefix + ": " + ec.prefix
			return
		}
		ec.prefix = prefix
	}
}

func WithStatusCode(code int) Option {
	return func(ec *ErrorContext) {
		ec.statusCode = code
	}
}

func WithPublicMessage(msg string) Option {
	return func(ec *ErrorContext) {
		ec.publicMsg = msg
	}
}

func WithShouldReport(report bool) Option {
	return func(ec *ErrorContext) {
		ec.shouldReport = report
	}
}

func WithDocsURL(url string) Option {
	return func(ec *ErrorContext) {
		ec.docsUrl = url
	}
}
