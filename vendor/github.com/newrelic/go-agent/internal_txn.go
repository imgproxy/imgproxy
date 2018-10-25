package newrelic

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/newrelic/go-agent/internal"
)

type txnInput struct {
	W          http.ResponseWriter
	Config     Config
	Reply      *internal.ConnectReply
	Consumer   dataConsumer
	attrConfig *internal.AttributeConfig
}

type txn struct {
	txnInput
	// This mutex is required since the consumer may call the public API
	// interface functions from different routines.
	sync.Mutex
	// finished indicates whether or not End() has been called.  After
	// finished has been set to true, no recording should occur.
	finished           bool
	numPayloadsCreated uint32

	ignore bool

	// wroteHeader prevents capturing multiple response code errors if the
	// user erroneously calls WriteHeader multiple times.
	wroteHeader bool

	internal.TxnData
}

func newTxn(input txnInput, req *http.Request, name string) *txn {
	txn := &txn{
		txnInput: input,
	}
	txn.Start = time.Now()
	txn.Name = name
	txn.IsWeb = nil != req
	txn.Attrs = internal.NewAttributes(input.attrConfig)

	if input.Config.DistributedTracer.Enabled {
		txn.BetterCAT.Enabled = true
		txn.BetterCAT.Priority = internal.NewPriority()
		txn.BetterCAT.ID = internal.NewSpanID()

		// Calculate sampled at the beginning of the transaction (rather
		// than lazily at payload creation time) because it controls the
		// creation of span events.
		txn.BetterCAT.Sampled = txn.Reply.AdaptiveSampler.ComputeSampled(txn.BetterCAT.Priority.Float32(), txn.Start)
		if txn.BetterCAT.Sampled {
			txn.BetterCAT.Priority += 1.0
		}
		txn.SpanEventsEnabled = input.Config.SpanEvents.Enabled
	}

	if nil != req {
		txn.Queuing = internal.QueueDuration(req.Header, txn.Start)
		internal.RequestAgentAttributes(txn.Attrs, req)
	}
	txn.Attrs.Agent.HostDisplayName = txn.Config.HostDisplayName
	txn.TxnTrace.Enabled = txn.txnTracesEnabled()
	txn.TxnTrace.SegmentThreshold = txn.Config.TransactionTracer.SegmentThreshold
	txn.StackTraceThreshold = txn.Config.TransactionTracer.StackTraceThreshold
	txn.SlowQueriesEnabled = txn.slowQueriesEnabled()
	txn.SlowQueryThreshold = txn.Config.DatastoreTracer.SlowQuery.Threshold
	if nil != req && nil != req.URL {
		txn.CleanURL = internal.SafeURL(req.URL)
	}

	// Synthetics support is tied up with a transaction's Old CAT field,
	// CrossProcess. To support Synthetics with either BetterCAT or Old CAT,
	// Initialize the CrossProcess field of the transaction, passing in
	// the top-level configuration.
	doOldCAT := txn.Config.CrossApplicationTracer.Enabled
	noGUID := txn.Config.DistributedTracer.Enabled
	txn.CrossProcess.InitFromHTTPRequest(doOldCAT, noGUID, input.Reply, req)

	return txn
}

func (txn *txn) slowQueriesEnabled() bool {
	return txn.Config.DatastoreTracer.SlowQuery.Enabled &&
		txn.Reply.CollectTraces
}

func (txn *txn) txnTracesEnabled() bool {
	return txn.Config.TransactionTracer.Enabled &&
		txn.Reply.CollectTraces
}

func (txn *txn) txnEventsEnabled() bool {
	return txn.Config.TransactionEvents.Enabled &&
		txn.Reply.CollectAnalyticsEvents
}

func (txn *txn) errorEventsEnabled() bool {
	return txn.Config.ErrorCollector.CaptureEvents &&
		txn.Reply.CollectErrorEvents
}

func (txn *txn) freezeName() {
	if txn.ignore || ("" != txn.FinalName) {
		return
	}

	txn.FinalName = internal.CreateFullTxnName(txn.Name, txn.Reply, txn.IsWeb)
	if "" == txn.FinalName {
		txn.ignore = true
	}
}

func (txn *txn) getsApdex() bool {
	return txn.IsWeb
}

func (txn *txn) txnTraceThreshold() time.Duration {
	if txn.Config.TransactionTracer.Threshold.IsApdexFailing {
		return internal.ApdexFailingThreshold(txn.ApdexThreshold)
	}
	return txn.Config.TransactionTracer.Threshold.Duration
}

func (txn *txn) shouldSaveTrace() bool {
	return txn.CrossProcess.IsSynthetics() ||
		(txn.txnTracesEnabled() && (txn.Duration >= txn.txnTraceThreshold()))
}

func (txn *txn) MergeIntoHarvest(h *internal.Harvest) {

	var priority internal.Priority
	if txn.BetterCAT.Enabled {
		priority = txn.BetterCAT.Priority
	} else {
		priority = internal.NewPriority()
	}

	internal.CreateTxnMetrics(&txn.TxnData, h.Metrics)
	internal.MergeBreakdownMetrics(&txn.TxnData, h.Metrics)

	if txn.txnEventsEnabled() {
		// Allocate a new TxnEvent to prevent a reference to the large transaction.
		alloc := new(internal.TxnEvent)
		*alloc = txn.TxnData.TxnEvent
		h.TxnEvents.AddTxnEvent(alloc, priority)
	}

	internal.MergeTxnErrors(&h.ErrorTraces, txn.Errors, txn.TxnEvent)

	if txn.errorEventsEnabled() {
		for _, e := range txn.Errors {
			errEvent := &internal.ErrorEvent{
				ErrorData: *e,
				TxnEvent:  txn.TxnEvent,
			}
			// Since the stack trace is not used in error events, remove the reference
			// to minimize memory.
			errEvent.Stack = nil
			h.ErrorEvents.Add(errEvent, priority)
		}
	}

	if txn.shouldSaveTrace() {
		h.TxnTraces.Witness(internal.HarvestTrace{
			TxnEvent: txn.TxnEvent,
			Trace:    txn.TxnTrace,
		})
	}

	if nil != txn.SlowQueries {
		h.SlowSQLs.Merge(txn.SlowQueries, txn.TxnEvent)
	}

	if txn.BetterCAT.Sampled && txn.Config.SpanEvents.Enabled {
		h.SpanEvents.MergeFromTransaction(&txn.TxnData)
	}
}

// TransportType's name field is not mutable outside of its package
// however, it still periodically needs to be used and assigned within
// the this package.  For testing purposes only.
func getTransport(transport string) string {
	var retVal string

	switch transport {
	case TransportHTTP.name:
		retVal = TransportHTTP.name
	case TransportHTTPS.name:
		retVal = TransportHTTPS.name
	case TransportKafka.name:
		retVal = TransportKafka.name
	case TransportJMS.name:
		retVal = TransportJMS.name
	case TransportIronMQ.name:
		retVal = TransportIronMQ.name
	case TransportAMQP.name:
		retVal = TransportAMQP.name
	case TransportQueue.name:
		retVal = TransportQueue.name
	case TransportOther.name:
		retVal = TransportOther.name
	case TransportUnknown.name:
	default:
		retVal = TransportUnknown.name
	}
	return retVal
}

func responseCodeIsError(cfg *Config, code int) bool {
	if code < http.StatusBadRequest { // 400
		return false
	}
	for _, ignoreCode := range cfg.ErrorCollector.IgnoreStatusCodes {
		if code == ignoreCode {
			return false
		}
	}
	return true
}

func headersJustWritten(txn *txn, code int) {
	txn.Lock()
	defer txn.Unlock()

	if txn.finished {
		return
	}
	if txn.wroteHeader {
		return
	}
	txn.wroteHeader = true

	internal.ResponseHeaderAttributes(txn.Attrs, txn.W.Header())
	internal.ResponseCodeAttribute(txn.Attrs, code)

	if responseCodeIsError(&txn.Config, code) {
		e := internal.TxnErrorFromResponseCode(time.Now(), code)
		e.Stack = internal.GetStackTrace(1)
		txn.noticeErrorInternal(e)
	}
}

func (txn *txn) responseHeader() http.Header {
	txn.Lock()
	defer txn.Unlock()

	if txn.finished {
		return nil
	}
	if txn.wroteHeader {
		return nil
	}
	if !txn.CrossProcess.Enabled {
		return nil
	}
	if !txn.CrossProcess.IsInbound() {
		return nil
	}
	txn.freezeName()
	contentLength := internal.GetContentLengthFromHeader(txn.W.Header())

	appData, err := txn.CrossProcess.CreateAppData(txn.FinalName, txn.Queuing, time.Since(txn.Start), contentLength)
	if err != nil {
		txn.Config.Logger.Debug("error generating outbound response header", map[string]interface{}{
			"error": err,
		})
		return nil
	}
	return internal.AppDataToHTTPHeader(appData)
}

func addCrossProcessHeaders(txn *txn) {
	// responseHeader() checks the wroteHeader field and returns a nil map if the
	// header has been written, so we don't need a check here.
	for key, values := range txn.responseHeader() {
		for _, value := range values {
			txn.W.Header().Add(key, value)
		}
	}
}

func (txn *txn) Header() http.Header { return txn.W.Header() }

func (txn *txn) Write(b []byte) (int, error) {
	// This is safe to call unconditionally, even if Write() is called multiple
	// times; see also the commentary in addCrossProcessHeaders().
	addCrossProcessHeaders(txn)

	n, err := txn.W.Write(b)

	headersJustWritten(txn, http.StatusOK)

	return n, err
}

func (txn *txn) WriteHeader(code int) {
	addCrossProcessHeaders(txn)

	txn.W.WriteHeader(code)

	headersJustWritten(txn, code)
}

func (txn *txn) End() error {
	txn.Lock()
	defer txn.Unlock()

	if txn.finished {
		return errAlreadyEnded
	}

	txn.finished = true

	r := recover()
	if nil != r {
		e := internal.TxnErrorFromPanic(time.Now(), r)
		e.Stack = internal.GetStackTrace(0)
		txn.noticeErrorInternal(e)
	}

	txn.Stop = time.Now()
	txn.Duration = txn.Stop.Sub(txn.Start)
	if children := internal.TracerRootChildren(&txn.TxnData); txn.Duration > children {
		txn.Exclusive = txn.Duration - children
	}

	txn.freezeName()

	// Finalise the CAT state.
	if err := txn.CrossProcess.Finalise(txn.Name, txn.Config.AppName); err != nil {
		txn.Config.Logger.Debug("error finalising the cross process state", map[string]interface{}{
			"error": err,
		})
	}

	// Assign apdexThreshold regardless of whether or not the transaction
	// gets apdex since it may be used to calculate the trace threshold.
	txn.ApdexThreshold = internal.CalculateApdexThreshold(txn.Reply, txn.FinalName)

	if txn.getsApdex() {
		if txn.HasErrors() {
			txn.Zone = internal.ApdexFailing
		} else {
			txn.Zone = internal.CalculateApdexZone(txn.ApdexThreshold, txn.Duration)
		}
	} else {
		txn.Zone = internal.ApdexNone
	}

	if txn.Config.Logger.DebugEnabled() {
		txn.Config.Logger.Debug("transaction ended", map[string]interface{}{
			"name":        txn.FinalName,
			"duration_ms": txn.Duration.Seconds() * 1000.0,
			"ignored":     txn.ignore,
			"run":         txn.Reply.RunID,
		})
	}

	if !txn.ignore {
		txn.Consumer.Consume(txn.Reply.RunID, txn)
	}

	// Note that if a consumer uses `panic(nil)`, the panic will not
	// propagate.
	if nil != r {
		panic(r)
	}

	return nil
}

func (txn *txn) AddAttribute(name string, value interface{}) error {
	txn.Lock()
	defer txn.Unlock()

	if txn.Config.HighSecurity {
		return errHighSecurityEnabled
	}

	if !txn.Reply.SecurityPolicies.CustomParameters.Enabled() {
		return errSecurityPolicy
	}

	if txn.finished {
		return errAlreadyEnded
	}

	return internal.AddUserAttribute(txn.Attrs, name, value, internal.DestAll)
}

var (
	errorsLocallyDisabled  = errors.New("errors locally disabled")
	errorsRemotelyDisabled = errors.New("errors remotely disabled")
	errNilError            = errors.New("nil error")
	errAlreadyEnded        = errors.New("transaction has already ended")
	errSecurityPolicy      = errors.New("disabled by security policy")
)

const (
	highSecurityErrorMsg   = "message removed by high security setting"
	securityPolicyErrorMsg = "message removed by security policy"
)

func (txn *txn) noticeErrorInternal(err internal.ErrorData) error {
	if !txn.Config.ErrorCollector.Enabled {
		return errorsLocallyDisabled
	}

	if !txn.Reply.CollectErrors {
		return errorsRemotelyDisabled
	}

	if nil == txn.Errors {
		txn.Errors = internal.NewTxnErrors(internal.MaxTxnErrors)
	}

	if txn.Config.HighSecurity {
		err.Msg = highSecurityErrorMsg
	}

	if !txn.Reply.SecurityPolicies.AllowRawExceptionMessages.Enabled() {
		err.Msg = securityPolicyErrorMsg
	}

	txn.Errors.Add(err)
	txn.TxnData.TxnEvent.HasError = true //mark transaction as having an error
	return nil
}

var (
	errTooManyErrorAttributes = fmt.Errorf("too many extra attributes: limit is %d",
		internal.AttributeErrorLimit)
)

func (txn *txn) NoticeError(err error) error {
	txn.Lock()
	defer txn.Unlock()

	if txn.finished {
		return errAlreadyEnded
	}

	if nil == err {
		return errNilError
	}

	e := internal.ErrorData{
		When: time.Now(),
		Msg:  err.Error(),
	}
	if ec, ok := err.(ErrorClasser); ok {
		e.Klass = ec.ErrorClass()
	}
	if "" == e.Klass {
		e.Klass = reflect.TypeOf(err).String()
	}
	if st, ok := err.(StackTracer); ok {
		e.Stack = st.StackTrace()
		// Note that if the provided stack trace is excessive in length,
		// it will be truncated during JSON creation.
	}
	if nil == e.Stack {
		e.Stack = internal.GetStackTrace(2)
	}

	if ea, ok := err.(ErrorAttributer); ok && !txn.Config.HighSecurity && txn.Reply.SecurityPolicies.CustomParameters.Enabled() {
		unvetted := ea.ErrorAttributes()
		if len(unvetted) > internal.AttributeErrorLimit {
			return errTooManyErrorAttributes
		}

		e.ExtraAttributes = make(map[string]interface{})
		for key, val := range unvetted {
			val, errr := internal.ValidateUserAttribute(key, val)
			if nil != errr {
				return errr
			}
			e.ExtraAttributes[key] = val
		}
	}

	return txn.noticeErrorInternal(e)
}

func (txn *txn) SetName(name string) error {
	txn.Lock()
	defer txn.Unlock()

	if txn.finished {
		return errAlreadyEnded
	}

	txn.Name = name
	return nil
}

func (txn *txn) Ignore() error {
	txn.Lock()
	defer txn.Unlock()

	if txn.finished {
		return errAlreadyEnded
	}
	txn.ignore = true
	return nil
}

func (txn *txn) StartSegmentNow() SegmentStartTime {
	var s internal.SegmentStartTime
	txn.Lock()
	if !txn.finished {
		s = internal.StartSegment(&txn.TxnData, time.Now())
	}
	txn.Unlock()
	return SegmentStartTime{
		segment: segment{
			start: s,
			txn:   txn,
		},
	}
}

type segment struct {
	start internal.SegmentStartTime
	txn   *txn
}

func endSegment(s *Segment) error {
	txn := s.StartTime.txn
	if nil == txn {
		return nil
	}
	var err error
	txn.Lock()
	if txn.finished {
		err = errAlreadyEnded
	} else {
		err = internal.EndBasicSegment(&txn.TxnData, s.StartTime.start, time.Now(), s.Name)
	}
	txn.Unlock()
	return err
}

func endDatastore(s *DatastoreSegment) error {
	txn := s.StartTime.txn
	if nil == txn {
		return nil
	}
	txn.Lock()
	defer txn.Unlock()

	if txn.finished {
		return errAlreadyEnded
	}
	if txn.Config.HighSecurity {
		s.QueryParameters = nil
	}
	if !txn.Config.DatastoreTracer.QueryParameters.Enabled {
		s.QueryParameters = nil
	}
	if txn.Reply.SecurityPolicies.RecordSQL.IsSet() {
		s.QueryParameters = nil
		if !txn.Reply.SecurityPolicies.RecordSQL.Enabled() {
			s.ParameterizedQuery = ""
		}
	}
	if !txn.Config.DatastoreTracer.DatabaseNameReporting.Enabled {
		s.DatabaseName = ""
	}
	if !txn.Config.DatastoreTracer.InstanceReporting.Enabled {
		s.Host = ""
		s.PortPathOrID = ""
	}
	return internal.EndDatastoreSegment(internal.EndDatastoreParams{
		Tracer:             &txn.TxnData,
		Start:              s.StartTime.start,
		Now:                time.Now(),
		Product:            string(s.Product),
		Collection:         s.Collection,
		Operation:          s.Operation,
		ParameterizedQuery: s.ParameterizedQuery,
		QueryParameters:    s.QueryParameters,
		Host:               s.Host,
		PortPathOrID:       s.PortPathOrID,
		Database:           s.DatabaseName,
	})
}

func externalSegmentMethod(s *ExternalSegment) string {
	r := s.Request

	// Is this a client request?
	if nil != s.Response && nil != s.Response.Request {
		r = s.Response.Request

		// Golang's http package states that when a client's
		// Request has an empty string for Method, the
		// method is GET.
		if "" == r.Method {
			return "GET"
		}
	}
	if nil == r {
		return ""
	}
	return r.Method
}

func externalSegmentURL(s *ExternalSegment) (*url.URL, error) {
	if "" != s.URL {
		return url.Parse(s.URL)
	}
	r := s.Request
	if nil != s.Response && nil != s.Response.Request {
		r = s.Response.Request
	}
	if r != nil {
		return r.URL, nil
	}
	return nil, nil
}

func endExternal(s *ExternalSegment) error {
	txn := s.StartTime.txn
	if nil == txn {
		return nil
	}
	txn.Lock()
	defer txn.Unlock()

	if txn.finished {
		return errAlreadyEnded
	}
	m := externalSegmentMethod(s)
	u, err := externalSegmentURL(s)
	if nil != err {
		return err
	}
	return internal.EndExternalSegment(&txn.TxnData, s.StartTime.start, time.Now(), u, m, s.Response)
}

// oldCATOutboundHeaders generates the Old CAT and Synthetics headers, depending
// on whether Old CAT is enabled or any Synthetics functionality has been
// triggered in the agent.
func oldCATOutboundHeaders(txn *txn) http.Header {
	txn.Lock()
	defer txn.Unlock()

	if txn.finished {
		return http.Header{}
	}

	metadata, err := txn.CrossProcess.CreateCrossProcessMetadata(txn.Name, txn.Config.AppName)
	if err != nil {
		txn.Config.Logger.Debug("error generating outbound headers", map[string]interface{}{
			"error": err,
		})

		// It's possible for CreateCrossProcessMetadata() to error and still have a
		// Synthetics header, so we'll still fall through to returning headers
		// based on whatever metadata was returned.
	}

	return internal.MetadataToHTTPHeader(metadata)
}

func outboundHeaders(s *ExternalSegment) http.Header {
	txn := s.StartTime.txn

	if nil == txn {
		return http.Header{}
	}

	hdr := oldCATOutboundHeaders(txn)

	// hdr may be empty, or it may contain headers.  If DistributedTracer
	// is enabled, add more to the existing hdr
	if p := txn.CreateDistributedTracePayload().HTTPSafe(); "" != p {
		hdr.Add(DistributedTracePayloadHeader, p)
		return hdr
	}

	return hdr
}

const (
	maxSampledDistributedPayloads = 35
)

type shimPayload struct{}

func (s shimPayload) Text() string     { return "" }
func (s shimPayload) HTTPSafe() string { return "" }

func (txn *txn) CreateDistributedTracePayload() (payload DistributedTracePayload) {
	payload = shimPayload{}

	txn.Lock()
	defer txn.Unlock()

	if !txn.BetterCAT.Enabled {
		return
	}

	if txn.finished {
		txn.CreatePayloadException = true
		return
	}

	if "" == txn.Reply.PrimaryAppID {
		// Return a shimPayload if the application is not yet connected.
		return
	}

	txn.numPayloadsCreated++

	var p internal.Payload
	p.Type = internal.CallerType
	p.Account = txn.Reply.AccountID

	p.App = txn.Reply.PrimaryAppID
	p.TracedID = txn.BetterCAT.TraceID()
	p.Priority = txn.BetterCAT.Priority
	p.Timestamp.Set(time.Now())
	p.TransactionID = txn.BetterCAT.ID // Set the transaction ID to the transaction guid.

	if txn.Reply.AccountID != txn.Reply.TrustedAccountKey {
		p.TrustedAccountKey = txn.Reply.TrustedAccountKey
	}

	if txn.BetterCAT.Sampled && txn.Config.SpanEvents.Enabled {
		p.ID = txn.CurrentSpanIdentifier()
	}

	// limit the number of outbound sampled=true payloads to prevent too
	// many downstream sampled events.
	p.SetSampled(false)
	if txn.numPayloadsCreated < maxSampledDistributedPayloads {
		p.SetSampled(txn.BetterCAT.Sampled)
	}

	txn.CreatePayloadSuccess = true

	payload = p
	return
}

var (
	errOutboundPayloadCreated   = errors.New("outbound payload already created")
	errAlreadyAccepted          = errors.New("AcceptDistributedTracePayload has already been called")
	errInboundPayloadDTDisabled = errors.New("DistributedTracer must be enabled to accept an inbound payload")
)

func (txn *txn) AcceptDistributedTracePayload(t TransportType, p interface{}) error {
	txn.Lock()
	defer txn.Unlock()

	if !txn.BetterCAT.Enabled {
		return errInboundPayloadDTDisabled
	}

	if txn.finished {
		txn.AcceptPayloadException = true
		return errAlreadyEnded
	}

	if txn.numPayloadsCreated > 0 {
		txn.AcceptPayloadCreateBeforeAccept = true
		return errOutboundPayloadCreated
	}

	if txn.BetterCAT.Inbound != nil {
		txn.AcceptPayloadIgnoredMultiple = true
		return errAlreadyAccepted
	}

	if nil == p {
		txn.AcceptPayloadNullPayload = true
		return nil
	}

	payload, err := internal.AcceptPayload(p)
	if nil != err {
		if _, ok := err.(internal.ErrPayloadParse); ok {
			txn.AcceptPayloadParseException = true
		} else if _, ok := err.(internal.ErrUnsupportedPayloadVersion); ok {
			txn.AcceptPayloadIgnoredVersion = true
		} else if _, ok := err.(internal.ErrPayloadMissingField); ok {
			txn.AcceptPayloadParseException = true
		} else {
			txn.AcceptPayloadException = true
		}
		return err
	}

	if nil == payload {
		return nil
	}

	// now that we have a parsed and alloc'd payload,
	// let's make  sure it has the correct fields
	if err := payload.IsValid(); nil != err {
		txn.AcceptPayloadParseException = true
		return err
	}

	// and let's also do our trustedKey check
	receivedTrustKey := payload.TrustedAccountKey
	if "" == receivedTrustKey {
		receivedTrustKey = payload.Account
	}
	if receivedTrustKey != txn.Reply.TrustedAccountKey {
		txn.AcceptPayloadUntrustedAccount = true
		return internal.ErrTrustedAccountKey{Message: "trusted account key missing or does not match"}
	}

	if 0 != payload.Priority {
		txn.BetterCAT.Priority = payload.Priority
	}

	// a nul payload.Sampled means the a field wasn't provided
	if nil != payload.Sampled {
		txn.BetterCAT.Sampled = *payload.Sampled
	} else {
		txn.BetterCAT.Sampled = txn.Reply.AdaptiveSampler.ComputeSampled(txn.BetterCAT.Priority.Float32(), time.Now())
	}

	txn.BetterCAT.Inbound = payload

	// TransportType's name field is not mutable outside of its package
	// so the only check needed is if the caller is using an empty TransportType
	txn.BetterCAT.Inbound.TransportType = t.name
	if t.name == "" {
		txn.BetterCAT.Inbound.TransportType = TransportUnknown.name
		txn.Config.Logger.Debug("Invalid transport type, defaulting to Unknown", map[string]interface{}{})
	}

	if tm := payload.Timestamp.Time(); txn.Start.After(tm) {
		txn.BetterCAT.Inbound.TransportDuration = txn.Start.Sub(tm)
	}

	txn.AcceptPayloadSuccess = true

	return nil
}
