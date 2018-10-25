package internal

import (
	"bytes"
	"time"
)

// https://source.datanerd.us/agents/agent-specs/blob/master/Span-Events.md

type spanCategory string

const (
	spanCategoryHTTP      spanCategory = "http"
	spanCategoryDatastore              = "datastore"
	spanCategoryGeneric                = "generic"
)

// SpanEvent represents a span event, neccessary to support Distributed Tracing.
type SpanEvent struct {
	TraceID         string
	GUID            string
	ParentID        string
	TransactionID   string
	Sampled         bool
	Priority        Priority
	Timestamp       time.Time
	Duration        time.Duration
	Name            string
	Category        spanCategory
	IsEntrypoint    bool
	DatastoreExtras *spanDatastoreExtras
	ExternalExtras  *spanExternalExtras
}

type spanDatastoreExtras struct {
	Component string
	Statement string
	Instance  string
	Address   string
	Hostname  string
}

type spanExternalExtras struct {
	URL       string
	Method    string
	Component string
}

// WriteJSON prepares JSON in the format expected by the collector.
func (e *SpanEvent) WriteJSON(buf *bytes.Buffer) {
	w := jsonFieldsWriter{buf: buf}
	buf.WriteByte('[')
	buf.WriteByte('{')
	w.stringField("type", "Span")
	w.stringField("traceId", e.TraceID)
	w.stringField("guid", e.GUID)
	if "" != e.ParentID {
		w.stringField("parentId", e.ParentID)
	}
	w.stringField("transactionId", e.TransactionID)
	w.boolField("sampled", e.Sampled)
	w.writerField("priority", e.Priority)
	w.intField("timestamp", e.Timestamp.UnixNano()/(1000*1000)) // in milliseconds
	w.floatField("duration", e.Duration.Seconds())
	w.stringField("name", e.Name)
	w.stringField("category", string(e.Category))
	if e.IsEntrypoint {
		w.boolField("nr.entryPoint", true)
	}
	if ex := e.DatastoreExtras; nil != ex {
		if "" != ex.Component {
			w.stringField("component", ex.Component)
		}
		if "" != ex.Statement {
			w.stringField("db.statement", ex.Statement)
		}
		if "" != ex.Instance {
			w.stringField("db.instance", ex.Instance)
		}
		if "" != ex.Address {
			w.stringField("peer.address", ex.Address)
		}
		if "" != ex.Hostname {
			w.stringField("peer.hostname", ex.Hostname)
		}
		w.stringField("span.kind", "client")
	}

	if ex := e.ExternalExtras; nil != ex {
		if "" != ex.URL {
			w.stringField("http.url", ex.URL)
		}
		if "" != ex.Method {
			w.stringField("http.method", ex.Method)
		}
		w.stringField("span.kind", "client")
		w.stringField("component", "http")
	}

	buf.WriteByte('}')
	buf.WriteByte(',')
	buf.WriteByte('{')
	buf.WriteByte('}')
	buf.WriteByte(',')
	buf.WriteByte('{')
	buf.WriteByte('}')
	buf.WriteByte(']')
}

// MarshalJSON is used for testing.
func (e *SpanEvent) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 256))

	e.WriteJSON(buf)

	return buf.Bytes(), nil
}

type spanEvents struct {
	events *analyticsEvents
}

func newSpanEvents(max int) *spanEvents {
	return &spanEvents{
		events: newAnalyticsEvents(max),
	}
}

func (events *spanEvents) addEvent(e *SpanEvent, cat *BetterCAT) {
	e.TraceID = cat.TraceID()
	e.TransactionID = cat.ID
	e.Sampled = cat.Sampled
	e.Priority = cat.Priority
	events.events.addEvent(analyticsEvent{priority: cat.Priority, jsonWriter: e})
}

// MergeFromTransaction merges the span events from a transaction into the
// harvest's span events.  This should only be called if the transaction was
// sampled and span events are enabled.
func (events *spanEvents) MergeFromTransaction(txndata *TxnData) {
	root := &SpanEvent{
		GUID:         txndata.getRootSpanID(),
		Timestamp:    txndata.Start,
		Duration:     txndata.Duration,
		Name:         txndata.FinalName,
		Category:     spanCategoryGeneric,
		IsEntrypoint: true,
	}
	if nil != txndata.BetterCAT.Inbound {
		root.ParentID = txndata.BetterCAT.Inbound.ID
	}
	events.addEvent(root, &txndata.BetterCAT)

	for _, evt := range txndata.spanEvents {
		events.addEvent(evt, &txndata.BetterCAT)
	}
}

func (events *spanEvents) MergeIntoHarvest(h *Harvest) {
	h.SpanEvents.events.mergeFailed(events.events)
}

func (events *spanEvents) Data(agentRunID string, harvestStart time.Time) ([]byte, error) {
	return events.events.CollectorJSON(agentRunID)
}

func (events *spanEvents) numSeen() float64  { return events.events.NumSeen() }
func (events *spanEvents) numSaved() float64 { return events.events.NumSaved() }

func (events *spanEvents) EndpointMethod() string {
	return cmdSpanEvents
}
