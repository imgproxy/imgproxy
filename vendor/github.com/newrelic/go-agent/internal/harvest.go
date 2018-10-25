package internal

import (
	"strings"
	"sync"
	"time"
)

// Harvestable is something that can be merged into a Harvest.
type Harvestable interface {
	MergeIntoHarvest(h *Harvest)
}

// Harvest contains collected data.
type Harvest struct {
	Metrics      *metricTable
	CustomEvents *customEvents
	TxnEvents    *txnEvents
	ErrorEvents  *errorEvents
	ErrorTraces  harvestErrors
	TxnTraces    *harvestTraces
	SlowSQLs     *slowQueries
	SpanEvents   *spanEvents
}

const (
	// txnEventPayloadlimit is the maximum number of events that should be
	// sent up in one post.
	txnEventPayloadlimit = 5000
)

// Payloads returns a map from expected collector method name to data type.
func (h *Harvest) Payloads(splitLargeTxnEvents bool) []PayloadCreator {
	ps := []PayloadCreator{
		h.Metrics,
		h.CustomEvents,
		h.ErrorEvents,
		h.ErrorTraces,
		h.TxnTraces,
		h.SlowSQLs,
		h.SpanEvents,
	}
	if splitLargeTxnEvents {
		ps = append(ps, h.TxnEvents.payloads(txnEventPayloadlimit)...)
	} else {
		ps = append(ps, h.TxnEvents)
	}
	return ps
}

// NewHarvest returns a new Harvest.
func NewHarvest(now time.Time) *Harvest {
	return &Harvest{
		Metrics:      newMetricTable(maxMetrics, now),
		CustomEvents: newCustomEvents(maxCustomEvents),
		TxnEvents:    newTxnEvents(maxTxnEvents),
		ErrorEvents:  newErrorEvents(maxErrorEvents),
		ErrorTraces:  newHarvestErrors(maxHarvestErrors),
		TxnTraces:    newHarvestTraces(),
		SlowSQLs:     newSlowQueries(maxHarvestSlowSQLs),
		SpanEvents:   newSpanEvents(maxSpanEvents),
	}
}

var (
	trackMutex   sync.Mutex
	trackMetrics []string
)

// TrackUsage helps track which integration packages are used.
func TrackUsage(s ...string) {
	trackMutex.Lock()
	defer trackMutex.Unlock()

	m := "Supportability/" + strings.Join(s, "/")
	trackMetrics = append(trackMetrics, m)
}

func createTrackUsageMetrics(metrics *metricTable) {
	trackMutex.Lock()
	defer trackMutex.Unlock()

	for _, m := range trackMetrics {
		metrics.addSingleCount(m, forced)
	}
}

// CreateFinalMetrics creates extra metrics at harvest time.
func (h *Harvest) CreateFinalMetrics() {
	h.Metrics.addSingleCount(instanceReporting, forced)

	h.Metrics.addCount(customEventsSeen, h.CustomEvents.numSeen(), forced)
	h.Metrics.addCount(customEventsSent, h.CustomEvents.numSaved(), forced)

	h.Metrics.addCount(txnEventsSeen, h.TxnEvents.numSeen(), forced)
	h.Metrics.addCount(txnEventsSent, h.TxnEvents.numSaved(), forced)

	h.Metrics.addCount(errorEventsSeen, h.ErrorEvents.numSeen(), forced)
	h.Metrics.addCount(errorEventsSent, h.ErrorEvents.numSaved(), forced)

	h.Metrics.addCount(spanEventsSeen, h.SpanEvents.numSeen(), forced)
	h.Metrics.addCount(spanEventsSent, h.SpanEvents.numSaved(), forced)

	if h.Metrics.numDropped > 0 {
		h.Metrics.addCount(supportabilityDropped, float64(h.Metrics.numDropped), forced)
	}

	createTrackUsageMetrics(h.Metrics)
}

// PayloadCreator is a data type in the harvest.
type PayloadCreator interface {
	// In the event of a rpm request failure (hopefully simply an
	// intermittent collector issue) the payload may be merged into the next
	// time period's harvest.
	Harvestable
	// Data prepares JSON in the format expected by the collector endpoint.
	// This method should return (nil, nil) if the payload is empty and no
	// rpm request is necessary.
	Data(agentRunID string, harvestStart time.Time) ([]byte, error)
	// EndpointMethod is used for the "method" query parameter when posting
	// the data.
	EndpointMethod() string
}

func supportMetric(metrics *metricTable, b bool, metricName string) {
	if b {
		metrics.addSingleCount(metricName, forced)
	}
}

// CreateTxnMetrics creates metrics for a transaction.
func CreateTxnMetrics(args *TxnData, metrics *metricTable) {
	// Duration Metrics
	rollup := backgroundRollup
	if args.IsWeb {
		rollup = webRollup
		metrics.addDuration(dispatcherMetric, "", args.Duration, 0, forced)
	}

	metrics.addDuration(args.FinalName, "", args.Duration, args.Exclusive, forced)
	metrics.addDuration(rollup, "", args.Duration, args.Exclusive, forced)

	// Better CAT Metrics
	if cat := args.BetterCAT; cat.Enabled {
		caller := callerUnknown
		if nil != cat.Inbound {
			caller = cat.Inbound.payloadCaller
		}
		m := durationByCallerMetric(caller)
		metrics.addDuration(m.all, "", args.Duration, args.Duration, unforced)
		metrics.addDuration(m.webOrOther(args.IsWeb), "", args.Duration, args.Duration, unforced)

		// Transport Duration Metric
		if nil != cat.Inbound {
			d := cat.Inbound.TransportDuration
			m = transportDurationMetric(caller)
			metrics.addDuration(m.all, "", d, d, unforced)
			metrics.addDuration(m.webOrOther(args.IsWeb), "", d, d, unforced)
		}

		// CAT Error Metrics
		if args.HasErrors() {
			m = errorsByCallerMetric(caller)
			metrics.addSingleCount(m.all, unforced)
			metrics.addSingleCount(m.webOrOther(args.IsWeb), unforced)
		}

		supportMetric(metrics, args.AcceptPayloadSuccess, supportTracingAcceptSuccess)
		supportMetric(metrics, args.AcceptPayloadException, supportTracingAcceptException)
		supportMetric(metrics, args.AcceptPayloadParseException, supportTracingAcceptParseException)
		supportMetric(metrics, args.AcceptPayloadCreateBeforeAccept, supportTracingCreateBeforeAccept)
		supportMetric(metrics, args.AcceptPayloadIgnoredMultiple, supportTracingIgnoredMultiple)
		supportMetric(metrics, args.AcceptPayloadIgnoredVersion, supportTracingIgnoredVersion)
		supportMetric(metrics, args.AcceptPayloadUntrustedAccount, supportTracingAcceptUntrustedAccount)
		supportMetric(metrics, args.AcceptPayloadNullPayload, supportTracingAcceptNull)
		supportMetric(metrics, args.CreatePayloadSuccess, supportTracingCreatePayloadSuccess)
		supportMetric(metrics, args.CreatePayloadException, supportTracingCreatePayloadException)
	}

	// Apdex Metrics
	if args.Zone != ApdexNone {
		metrics.addApdex(apdexRollup, "", args.ApdexThreshold, args.Zone, forced)

		mname := apdexPrefix + removeFirstSegment(args.FinalName)
		metrics.addApdex(mname, "", args.ApdexThreshold, args.Zone, unforced)
	}

	// Error Metrics
	if args.HasErrors() {
		metrics.addSingleCount(errorsRollupMetric.all, forced)
		metrics.addSingleCount(errorsRollupMetric.webOrOther(args.IsWeb), forced)
		metrics.addSingleCount(errorsPrefix+args.FinalName, forced)
	}

	// Queueing Metrics
	if args.Queuing > 0 {
		metrics.addDuration(queueMetric, "", args.Queuing, args.Queuing, forced)
	}
}
