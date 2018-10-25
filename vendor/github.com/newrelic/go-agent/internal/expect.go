package internal

import (
	"encoding/json"
	"fmt"
	"runtime"
)

var (
	// Unfortunately, the resolution of time.Now() on Windows is coarse: Two
	// sequential calls to time.Now() may return the same value, and tests
	// which expect non-zero durations may fail.  To avoid adding sleep
	// statements or mocking time.Now(), those tests are skipped on Windows.
	doDurationTests = runtime.GOOS != `windows`
)

// Validator is used for testing.
type Validator interface {
	Error(...interface{})
}

func validateStringField(v Validator, fieldName, v1, v2 string) {
	if v1 != v2 {
		v.Error(fieldName, v1, v2)
	}
}

type addValidatorField struct {
	field    interface{}
	original Validator
}

func (a addValidatorField) Error(fields ...interface{}) {
	fields = append([]interface{}{a.field}, fields...)
	a.original.Error(fields...)
}

// ExtendValidator is used to add more context to a validator.
func ExtendValidator(v Validator, field interface{}) Validator {
	return addValidatorField{
		field:    field,
		original: v,
	}
}

// WantMetric is a metric expectation.  If Data is nil, then any data values are
// acceptable.
type WantMetric struct {
	Name   string
	Scope  string
	Forced interface{} // true, false, or nil
	Data   []float64
}

// WantError is a traced error expectation.
type WantError struct {
	TxnName         string
	Msg             string
	Klass           string
	Caller          string
	URL             string
	UserAttributes  map[string]interface{}
	AgentAttributes map[string]interface{}
}

func uniquePointer() *struct{} {
	s := struct{}{}
	return &s
}

var (
	// MatchAnything is for use when matching attributes.
	MatchAnything = uniquePointer()
)

// WantEvent is a transaction or error event expectation.
type WantEvent struct {
	Intrinsics      map[string]interface{}
	UserAttributes  map[string]interface{}
	AgentAttributes map[string]interface{}
}

// WantTxnTrace is a transaction trace expectation.
type WantTxnTrace struct {
	MetricName      string
	CleanURL        string
	NumSegments     int
	UserAttributes  map[string]interface{}
	AgentAttributes map[string]interface{}
}

// WantSlowQuery is a slowQuery expectation.
type WantSlowQuery struct {
	Count        int32
	MetricName   string
	Query        string
	TxnName      string
	TxnURL       string
	DatabaseName string
	Host         string
	PortPathOrID string
	Params       map[string]interface{}
}

// Expect exposes methods that allow for testing whether the correct data was
// captured.
type Expect interface {
	ExpectCustomEvents(t Validator, want []WantEvent)
	ExpectErrors(t Validator, want []WantError)
	ExpectErrorEvents(t Validator, want []WantEvent)
	ExpectErrorEventsPresent(t Validator, want []WantEvent)
	ExpectErrorEventsAbsent(t Validator, names []string)

	ExpectTxnEvents(t Validator, want []WantEvent)
	ExpectTxnEventsPresent(t Validator, want []WantEvent)
	ExpectTxnEventsAbsent(t Validator, names []string)

	ExpectMetrics(t Validator, want []WantMetric)
	ExpectMetricsPresent(t Validator, want []WantMetric)

	ExpectTxnTraces(t Validator, want []WantTxnTrace)
	ExpectSlowQueries(t Validator, want []WantSlowQuery)

	ExpectSpanEvents(t Validator, want []WantEvent)
	ExpectSpanEventsPresent(t Validator, want []WantEvent)
	ExpectSpanEventsAbsent(t Validator, names []string)
	ExpectSpanEventsCount(t Validator, c int)
}

func expectMetricField(t Validator, id metricID, v1, v2 float64, fieldName string) {
	if v1 != v2 {
		t.Error("metric fields do not match", id, v1, v2, fieldName)
	}
}

// ExpectMetricsPresent allows testing of metrics with requiring an exact match
func ExpectMetricsPresent(t Validator, mt *metricTable, expect []WantMetric) {
	expectedIds := make(map[metricID]struct{})
	for _, e := range expect {
		id := metricID{Name: e.Name, Scope: e.Scope}
		expectedIds[id] = struct{}{}
		m := mt.metrics[id]
		if nil == m {
			t.Error("unable to find metric", id)
			continue
		}

		if b, ok := e.Forced.(bool); ok {
			if b != (forced == m.forced) {
				t.Error("metric forced incorrect", b, m.forced, id)
			}
		}

		if nil != e.Data {
			expectMetricField(t, id, e.Data[0], m.data.countSatisfied, "countSatisfied")
			expectMetricField(t, id, e.Data[1], m.data.totalTolerated, "totalTolerated")
			expectMetricField(t, id, e.Data[2], m.data.exclusiveFailed, "exclusiveFailed")
			expectMetricField(t, id, e.Data[3], m.data.min, "min")
			expectMetricField(t, id, e.Data[4], m.data.max, "max")
			expectMetricField(t, id, e.Data[5], m.data.sumSquares, "sumSquares")
		}
	}
}

// ExpectMetrics allows testing of metrics.  It passes if mt exactly matches expect.
func ExpectMetrics(t Validator, mt *metricTable, expect []WantMetric) {
	if len(mt.metrics) != len(expect) {
		t.Error("metric counts do not match expectations", len(mt.metrics), len(expect))
	}
	expectedIds := make(map[metricID]struct{})
	for _, e := range expect {
		id := metricID{Name: e.Name, Scope: e.Scope}
		expectedIds[id] = struct{}{}
		m := mt.metrics[id]
		if nil == m {
			t.Error("unable to find metric", id)
			continue
		}

		if b, ok := e.Forced.(bool); ok {
			if b != (forced == m.forced) {
				t.Error("metric forced incorrect", b, m.forced, id)
			}
		}

		if nil != e.Data {
			expectMetricField(t, id, e.Data[0], m.data.countSatisfied, "countSatisfied")
			expectMetricField(t, id, e.Data[1], m.data.totalTolerated, "totalTolerated")
			expectMetricField(t, id, e.Data[2], m.data.exclusiveFailed, "exclusiveFailed")
			expectMetricField(t, id, e.Data[3], m.data.min, "min")
			expectMetricField(t, id, e.Data[4], m.data.max, "max")
			expectMetricField(t, id, e.Data[5], m.data.sumSquares, "sumSquares")
		}
	}
	for id := range mt.metrics {
		if _, ok := expectedIds[id]; !ok {
			t.Error("expected metrics does not contain", id.Name, id.Scope)
		}
	}
}

func expectAttributesPresent(v Validator, exists map[string]interface{}, expect map[string]interface{}) {
	for key, val := range expect {
		found, ok := exists[key]
		if !ok {
			v.Error("expected attribute not found: ", key)
			continue
		}
		if val == MatchAnything {
			continue
		}
		v1 := fmt.Sprint(found)
		v2 := fmt.Sprint(val)
		if v1 != v2 {
			v.Error("value difference", fmt.Sprintf("key=%s", key), v1, v2)
		}
	}
}

func expectAttributes(v Validator, exists map[string]interface{}, expect map[string]interface{}) {
	// TODO: This params comparison can be made smarter: Alert differences
	// based on sub/super set behavior.
	if len(exists) != len(expect) {
		v.Error("attributes length difference", len(exists), len(expect))
	}
	for key, val := range expect {
		found, ok := exists[key]
		if !ok {
			v.Error("expected attribute not found: ", key)
			continue
		}
		if val == MatchAnything {
			continue
		}
		v1 := fmt.Sprint(found)
		v2 := fmt.Sprint(val)
		if v1 != v2 {
			v.Error("value difference", fmt.Sprintf("key=%s", key), v1, v2)
		}
	}
	for key, val := range exists {
		_, ok := expect[key]
		if !ok {
			v.Error("unexpected attribute present: ", key, val)
			continue
		}
	}
}

// ExpectCustomEvents allows testing of custom events.  It passes if cs exactly matches expect.
func ExpectCustomEvents(v Validator, cs *customEvents, expect []WantEvent) {
	if len(cs.events.events) != len(expect) {
		v.Error("number of custom events does not match", len(cs.events.events),
			len(expect))
		return
	}
	for i, e := range expect {
		event, ok := cs.events.events[i].jsonWriter.(*CustomEvent)
		if !ok {
			v.Error("wrong custom event")
		} else {
			expectEvent(v, event, e)
		}
	}
}

func expectEventAbsent(v Validator, e json.Marshaler, names []string) {
	js, err := e.MarshalJSON()
	if nil != err {
		v.Error("unable to marshal event", err)
		return
	}

	var event []map[string]interface{}
	err = json.Unmarshal(js, &event)
	if nil != err {
		v.Error("unable to parse event json", err)
		return
	}

	intrinsics := event[0]
	userAttributes := event[1]
	agentAttributes := event[2]

	for _, name := range names {
		if _, ok := intrinsics[name]; ok {
			v.Error("unexpected key found", name)
		}

		if _, ok := userAttributes[name]; ok {
			v.Error("unexpected key found", name)
		}

		if _, ok := agentAttributes[name]; ok {
			v.Error("unexpected key found", name)
		}
	}
}

func expectEventPresent(v Validator, e json.Marshaler, expect WantEvent) {
	js, err := e.MarshalJSON()
	if nil != err {
		v.Error("unable to marshal event", err)
		return
	}
	var event []map[string]interface{}
	err = json.Unmarshal(js, &event)
	if nil != err {
		v.Error("unable to parse event json", err)
		return
	}
	intrinsics := event[0]
	userAttributes := event[1]
	agentAttributes := event[2]

	if nil != expect.Intrinsics {
		expectAttributesPresent(v, intrinsics, expect.Intrinsics)
	}
	if nil != expect.UserAttributes {
		expectAttributesPresent(v, userAttributes, expect.UserAttributes)
	}
	if nil != expect.AgentAttributes {
		expectAttributesPresent(v, agentAttributes, expect.AgentAttributes)
	}
}

func expectEvent(v Validator, e json.Marshaler, expect WantEvent) {
	js, err := e.MarshalJSON()
	if nil != err {
		v.Error("unable to marshal event", err)
		return
	}
	var event []map[string]interface{}
	err = json.Unmarshal(js, &event)
	if nil != err {
		v.Error("unable to parse event json", err)
		return
	}
	intrinsics := event[0]
	userAttributes := event[1]
	agentAttributes := event[2]

	if nil != expect.Intrinsics {
		expectAttributes(v, intrinsics, expect.Intrinsics)
	}
	if nil != expect.UserAttributes {
		expectAttributes(v, userAttributes, expect.UserAttributes)
	}
	if nil != expect.AgentAttributes {
		expectAttributes(v, agentAttributes, expect.AgentAttributes)
	}
}

// Second attributes have priority.
func mergeAttributes(a1, a2 map[string]interface{}) map[string]interface{} {
	a := make(map[string]interface{})
	for k, v := range a1 {
		a[k] = v
	}
	for k, v := range a2 {
		a[k] = v
	}
	return a
}

// ExpectErrorEventsPresent allows testing of events with requiring an exact match
func ExpectErrorEventsPresent(v Validator, events *errorEvents, expect []WantEvent) {
	for i, e := range expect {
		event, ok := events.events.events[i].jsonWriter.(*ErrorEvent)
		if !ok {
			v.Error("wrong span event in ExpectErrorEventsPresent")
		} else {
			expectEventPresent(v, event, e)
		}
	}
}

// ExpectErrorEventsAbsent allows testing that a set of attribute names are absent from the event data
func ExpectErrorEventsAbsent(v Validator, events *errorEvents, names []string) {
	for _, eventHarvested := range events.events.events {
		event, ok := eventHarvested.jsonWriter.(*ErrorEvent)
		if !ok {
			v.Error("wrong span event in ExpectErrorEventsAbsent")
		} else {
			expectEventAbsent(v, event, names)
		}
	}
}

// ExpectErrorEvents allows testing of error events.  It passes if events exactly matches expect.
func ExpectErrorEvents(v Validator, events *errorEvents, expect []WantEvent) {
	if len(events.events.events) != len(expect) {
		v.Error("number of custom events does not match",
			len(events.events.events), len(expect))
		return
	}
	for i, e := range expect {
		event, ok := events.events.events[i].jsonWriter.(*ErrorEvent)
		if !ok {
			v.Error("wrong error event")
		} else {
			if nil != e.Intrinsics {
				e.Intrinsics = mergeAttributes(map[string]interface{}{
					// The following intrinsics should always be present in
					// error events:
					"type":      "TransactionError",
					"timestamp": MatchAnything,
					"duration":  MatchAnything,
				}, e.Intrinsics)
			}
			expectEvent(v, event, e)
		}
	}
}

// ExpectSpanEventsCount allows us to count how many events the system generated
func ExpectSpanEventsCount(v Validator, events *spanEvents, c int) {
	len := len(events.events.events)
	if len != c {
		v.Error(fmt.Sprintf("expected %d span events, found %d", c, len))
	}
}

// ExpectSpanEventsPresent allows us to test for the presence and value of events
// without also requiring an exact match
func ExpectSpanEventsPresent(v Validator, events *spanEvents, expect []WantEvent) {
	for i, e := range expect {
		event, ok := events.events.events[i].jsonWriter.(*SpanEvent)
		if !ok {
			v.Error("wrong span event in ExpectSpanEventsPresent")
		} else {
			expectEventPresent(v, event, e)
		}
	}
}

// ExpectSpanEventsAbsent allows us to ensure that a set of attribute names are absent
// from the event data
func ExpectSpanEventsAbsent(v Validator, events *spanEvents, names []string) {
	for _, eventHarvested := range events.events.events {
		event, ok := eventHarvested.jsonWriter.(*SpanEvent)
		if !ok {
			v.Error("wrong span event in ExpectSpanEventsAbsent")
		} else {
			expectEventAbsent(v, event, names)
		}
	}
}

// ExpectSpanEvents allows testing of span events.  It passes if events exactly matches expect.
func ExpectSpanEvents(v Validator, events *spanEvents, expect []WantEvent) {
	if len(events.events.events) != len(expect) {
		v.Error("number of txn events does not match",
			len(events.events.events), len(expect))
		return
	}
	for i, e := range expect {
		event, ok := events.events.events[i].jsonWriter.(*SpanEvent)
		if !ok {
			v.Error("wrong span event")
		} else {
			if nil != e.Intrinsics {
				e.Intrinsics = mergeAttributes(map[string]interface{}{
					// The following intrinsics should always be present in
					// span events:
					"type":      "Transaction",
					"timestamp": MatchAnything,
					"duration":  MatchAnything,
				}, e.Intrinsics)
			}
			expectEvent(v, event, e)
		}
	}
}

// ExpectTxnEventsPresent allows us to test for the presence and value of events
// without also requiring an exact match
func ExpectTxnEventsPresent(v Validator, events *txnEvents, expect []WantEvent) {
	for i, e := range expect {
		event, ok := events.events.events[i].jsonWriter.(*TxnEvent)
		if !ok {
			v.Error("wrong txn event in ExpectTxnEventsPresent")
		} else {
			expectEventPresent(v, event, e)
		}
	}
}

// ExpectTxnEventsAbsent allows us to ensure that a set of attribute names are absent
// from the event data
func ExpectTxnEventsAbsent(v Validator, events *txnEvents, names []string) {
	for _, eventHarvested := range events.events.events {
		event, ok := eventHarvested.jsonWriter.(*TxnEvent)
		if !ok {
			v.Error("wrong txn event in ExpectTxnEventsAbsent")
		} else {
			expectEventAbsent(v, event, names)
		}
	}
}

// ExpectTxnEvents allows testing of txn events.
func ExpectTxnEvents(v Validator, events *txnEvents, expect []WantEvent) {
	if len(events.events.events) != len(expect) {
		v.Error("number of txn events does not match",
			len(events.events.events), len(expect))
		return
	}
	for i, e := range expect {
		event, ok := events.events.events[i].jsonWriter.(*TxnEvent)
		if !ok {
			v.Error("wrong txn event")
		} else {
			if nil != e.Intrinsics {
				e.Intrinsics = mergeAttributes(map[string]interface{}{
					// The following intrinsics should always be present in
					// txn events:
					"type":      "Transaction",
					"timestamp": MatchAnything,
					"duration":  MatchAnything,
					"error":     MatchAnything,
				}, e.Intrinsics)
			}
			expectEvent(v, event, e)
		}
	}
}

func expectError(v Validator, err *tracedError, expect WantError) {
	caller := topCallerNameBase(err.ErrorData.Stack)
	validateStringField(v, "caller", expect.Caller, caller)
	validateStringField(v, "txnName", expect.TxnName, err.FinalName)
	validateStringField(v, "klass", expect.Klass, err.Klass)
	validateStringField(v, "msg", expect.Msg, err.Msg)
	validateStringField(v, "URL", expect.URL, err.CleanURL)
	js, errr := err.MarshalJSON()
	if nil != errr {
		v.Error("unable to marshal error json", errr)
		return
	}
	var unmarshalled []interface{}
	errr = json.Unmarshal(js, &unmarshalled)
	if nil != errr {
		v.Error("unable to unmarshal error json", errr)
		return
	}
	attributes := unmarshalled[4].(map[string]interface{})
	agentAttributes := attributes["agentAttributes"].(map[string]interface{})
	userAttributes := attributes["userAttributes"].(map[string]interface{})

	if nil != expect.UserAttributes {
		expectAttributes(v, userAttributes, expect.UserAttributes)
	}
	if nil != expect.AgentAttributes {
		expectAttributes(v, agentAttributes, expect.AgentAttributes)
	}
}

// ExpectErrors allows testing of errors.
func ExpectErrors(v Validator, errors harvestErrors, expect []WantError) {
	if len(errors) != len(expect) {
		v.Error("number of errors mismatch", len(errors), len(expect))
		return
	}
	for i, e := range expect {
		expectError(v, errors[i], e)
	}
}

func countSegments(node []interface{}) int {
	count := 1
	children := node[4].([]interface{})
	for _, c := range children {
		node := c.([]interface{})
		count += countSegments(node)
	}
	return count
}

func expectTxnTrace(v Validator, got json.Marshaler, expect WantTxnTrace) {
	js, err := got.MarshalJSON()
	if nil != err {
		v.Error("unable to marshal txn trace json", err)
		return
	}
	var unmarshalled []interface{}
	err = json.Unmarshal(js, &unmarshalled)
	if nil != err {
		v.Error("unable to unmarshal error json", err)
		return
	}
	duration := unmarshalled[1].(float64)
	name := unmarshalled[2].(string)
	cleanURL := unmarshalled[3].(string)
	traceData := unmarshalled[4].([]interface{})

	rootNode := traceData[3].([]interface{})
	attributes := traceData[4].(map[string]interface{})
	userAttributes := attributes["userAttributes"].(map[string]interface{})
	agentAttributes := attributes["agentAttributes"].(map[string]interface{})

	validateStringField(v, "metric name", expect.MetricName, name)
	validateStringField(v, "request url", expect.CleanURL, cleanURL)

	if doDurationTests && 0 == duration {
		v.Error("zero trace duration")
	}

	if nil != expect.UserAttributes {
		expectAttributes(v, userAttributes, expect.UserAttributes)
	}
	if nil != expect.AgentAttributes {
		expectAttributes(v, agentAttributes, expect.AgentAttributes)
	}
	numSegments := countSegments(rootNode)
	// The expectation segment count does not include the two root nodes.
	numSegments -= 2
	if expect.NumSegments != numSegments {
		v.Error("wrong number of segments", expect.NumSegments, numSegments)
	}
}

// ExpectTxnTraces allows testing of transaction traces.
func ExpectTxnTraces(v Validator, traces *harvestTraces, want []WantTxnTrace) {
	if len(want) != traces.Len() {
		v.Error("number of traces do not match", len(want), traces.Len())
	}

	actual := traces.slice()
	for i, expected := range want {
		expectTxnTrace(v, actual[i], expected)
	}
}

func expectSlowQuery(t Validator, slowQuery *slowQuery, want WantSlowQuery) {
	if slowQuery.Count != want.Count {
		t.Error("wrong Count field", slowQuery.Count, want.Count)
	}
	validateStringField(t, "MetricName", slowQuery.DatastoreMetric, want.MetricName)
	validateStringField(t, "Query", slowQuery.ParameterizedQuery, want.Query)
	validateStringField(t, "TxnEvent.FinalName", slowQuery.TxnEvent.FinalName, want.TxnName)
	validateStringField(t, "TxnEvent.CleanURL", slowQuery.TxnEvent.CleanURL, want.TxnURL)
	validateStringField(t, "DatabaseName", slowQuery.DatabaseName, want.DatabaseName)
	validateStringField(t, "Host", slowQuery.Host, want.Host)
	validateStringField(t, "PortPathOrID", slowQuery.PortPathOrID, want.PortPathOrID)
	expectAttributes(t, map[string]interface{}(slowQuery.QueryParameters), want.Params)
}

// ExpectSlowQueries allows testing of slow queries.
func ExpectSlowQueries(t Validator, slowQueries *slowQueries, want []WantSlowQuery) {
	if len(want) != len(slowQueries.priorityQueue) {
		t.Error("wrong number of slow queries",
			"expected", len(want), "got", len(slowQueries.priorityQueue))
		return
	}
	for _, s := range want {
		idx, ok := slowQueries.lookup[s.Query]
		if !ok {
			t.Error("unable to find slow query", s.Query)
			continue
		}
		expectSlowQuery(t, slowQueries.priorityQueue[idx], s)
	}
}
