package logging

import "context"

const (
	PeerIdContextLabel       = "ctx.peer_id"
	ConnectionIdContextLabel = "ctx.connection_id"

	// StreamIdContextLabel is used to store stream ids which are request numbers interpreted so that
	// streams initiated by us have positive numbers and streams initiated by remote have negative numbers.
	StreamIdContextLabel = "ctx.stream_id"
)

type loggingContextKeyType string

const loggingContextKey loggingContextKeyType = "logging_context"

type LoggingContext map[string]any

func AddToLoggingContext(ctx context.Context, label string, value any) context.Context {
	loggingContext := copyLoggingContext(GetLoggingContext(ctx))
	loggingContext[label] = value
	return context.WithValue(ctx, loggingContextKey, loggingContext)
}

func GetLoggingContext(ctx context.Context) LoggingContext {
	v := ctx.Value(loggingContextKey)
	if v == nil {
		return make(LoggingContext)
	}
	return v.(LoggingContext)
}

func copyLoggingContext(loggingContext LoggingContext) LoggingContext {
	result := make(LoggingContext, len(loggingContext))
	for key, value := range loggingContext {
		result[key] = value
	}
	return result
}
