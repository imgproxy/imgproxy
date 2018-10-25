package newrelic

import (
	"net/http"
)

// Transaction represents a request or a background task.
// Each Transaction should only be used in a single goroutine.
type Transaction interface {
	// If StartTransaction is called with a non-nil http.ResponseWriter then
	// the Transaction may be used in its place.  This allows
	// instrumentation of the response code and response headers.
	http.ResponseWriter

	// End finishes the current transaction, stopping all further
	// instrumentation.  Subsequent calls to End will have no effect.
	End() error

	// Ignore ensures that this transaction's data will not be recorded.
	Ignore() error

	// SetName names the transaction.  Transactions will not be grouped
	// usefully if too many unique names are used.
	SetName(name string) error

	// NoticeError records an error.  The first five errors per transaction
	// are recorded (this behavior is subject to potential change in the
	// future).
	NoticeError(err error) error

	// AddAttribute adds a key value pair to the current transaction.  This
	// information is attached to errors, transaction events, and error
	// events.  The key must contain fewer than than 255 bytes.  The value
	// must be a number, string, or boolean.  Attribute configuration is
	// applied (see config.go).
	//
	// For more information, see:
	// https://docs.newrelic.com/docs/agents/manage-apm-agents/agent-metrics/collect-custom-attributes
	AddAttribute(key string, value interface{}) error

	// StartSegmentNow allows the timing of functions, external calls, and
	// datastore calls.  The segments of each transaction MUST be used in a
	// single goroutine.  Consumers are encouraged to use the
	// `StartSegmentNow` functions which checks if the Transaction is nil.
	// See segments.go
	StartSegmentNow() SegmentStartTime

	// CreateDistributedTracePayload creates a payload to link the calls
	// between transactions. This method never returns nil. Instead, it may
	// return a shim implementation whose methods return empty strings.
	// CreateDistributedTracePayload should be called every time an outbound
	// call is made since the payload contains a timestamp.
	//
	// StartExternalSegment calls CreateDistributedTracePayload, so you
	// should not need to use this method for typical outbound HTTP calls.
	// Just use StartExternalSegment!
	CreateDistributedTracePayload() DistributedTracePayload

	// AcceptDistributedTracePayload is used at the beginning of a
	// transaction to identify the caller.
	//
	// Application.StartTransaction calls this method automatically if a
	// payload is present in the request headers (under the key
	// DistributedTracePayloadHeader).  Therefore, this method does not need
	// to be used for typical HTTP transactions.
	//
	// AcceptDistributedTracePayload should be used as early in the
	// transaction as possible. It may not be called after a call to
	// CreateDistributedTracePayload.
	//
	// The payload parameter may be a DistributedTracePayload or a string.
	AcceptDistributedTracePayload(t TransportType, payload interface{}) error
}

// DistributedTracePayload is used to instrument connections between
// transactions and applications.
type DistributedTracePayload interface {
	// HTTPSafe serializes the payload into a string containing http safe
	// characters.
	HTTPSafe() string
	// Text serializes the payload into a string.  The format is slightly
	// more compact than HTTPSafe.
	Text() string
}

const (
	// DistributedTracePayloadHeader is the header used by New Relic agents
	// for automatic trace payload instrumentation.
	DistributedTracePayloadHeader = "Newrelic"
)

// TransportType represents the type of connection that the trace payload was
// transported over.
type TransportType struct{ name string }

// TransportType names used across New Relic agents:
var (
	TransportUnknown = TransportType{name: "Unknown"}
	TransportHTTP    = TransportType{name: "HTTP"}
	TransportHTTPS   = TransportType{name: "HTTPS"}
	TransportKafka   = TransportType{name: "Kafka"}
	TransportJMS     = TransportType{name: "JMS"}
	TransportIronMQ  = TransportType{name: "IronMQ"}
	TransportAMQP    = TransportType{name: "AMQP"}
	TransportQueue   = TransportType{name: "Queue"}
	TransportOther   = TransportType{name: "Other"}
)
