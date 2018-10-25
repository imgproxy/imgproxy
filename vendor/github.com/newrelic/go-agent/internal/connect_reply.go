package internal

import (
	"encoding/json"
	"strings"
	"time"
)

// AgentRunID identifies the current connection with the collector.
type AgentRunID string

func (id AgentRunID) String() string {
	return string(id)
}

// PreconnectReply contains settings from the preconnect endpoint.
type PreconnectReply struct {
	Collector        string           `json:"redirect_host"`
	SecurityPolicies SecurityPolicies `json:"security_policies"`
}

// ConnectReply contains all of the settings and state send down from the
// collector.  It should not be modified after creation.
type ConnectReply struct {
	RunID AgentRunID `json:"agent_run_id"`

	// Transaction Name Modifiers
	SegmentTerms segmentRules `json:"transaction_segment_terms"`
	TxnNameRules metricRules  `json:"transaction_name_rules"`
	URLRules     metricRules  `json:"url_rules"`
	MetricRules  metricRules  `json:"metric_name_rules"`

	// Cross Process
	EncodingKey     string            `json:"encoding_key"`
	CrossProcessID  string            `json:"cross_process_id"`
	TrustedAccounts trustedAccountSet `json:"trusted_account_ids"`

	// Settings
	KeyTxnApdex            map[string]float64 `json:"web_transactions_apdex"`
	ApdexThresholdSeconds  float64            `json:"apdex_t"`
	CollectAnalyticsEvents bool               `json:"collect_analytics_events"`
	CollectCustomEvents    bool               `json:"collect_custom_events"`
	CollectTraces          bool               `json:"collect_traces"`
	CollectErrors          bool               `json:"collect_errors"`
	CollectErrorEvents     bool               `json:"collect_error_events"`

	// RUM
	AgentLoader string `json:"js_agent_loader"`
	Beacon      string `json:"beacon"`
	BrowserKey  string `json:"browser_key"`
	AppID       string `json:"application_id"`
	ErrorBeacon string `json:"error_beacon"`
	JSAgentFile string `json:"js_agent_file"`

	// PreconnectReply fields are not in the connect reply, this embedding
	// is done to simplify code.
	PreconnectReply `json:"-"`

	Messages []struct {
		Message string `json:"message"`
		Level   string `json:"level"`
	} `json:"messages"`

	AdaptiveSampler AdaptiveSampler

	// BetterCAT/Distributed Tracing
	AccountID                     string `json:"account_id"`
	TrustedAccountKey             string `json:"trusted_account_key"`
	PrimaryAppID                  string `json:"primary_application_id"`
	SamplingTarget                uint64 `json:"sampling_target"`
	SamplingTargetPeriodInSeconds int    `json:"sampling_target_period_in_seconds"`
}

type trustedAccountSet map[int]struct{}

func (t *trustedAccountSet) IsTrusted(account int) bool {
	_, exists := (*t)[account]
	return exists
}

func (t *trustedAccountSet) UnmarshalJSON(data []byte) error {
	accounts := make([]int, 0)
	if err := json.Unmarshal(data, &accounts); err != nil {
		return err
	}

	*t = make(trustedAccountSet)
	for _, account := range accounts {
		(*t)[account] = struct{}{}
	}

	return nil
}

// ConnectReplyDefaults returns a newly allocated ConnectReply with the proper
// default settings.  A pointer to a global is not used to prevent consumers
// from changing the default settings.
func ConnectReplyDefaults() *ConnectReply {
	return &ConnectReply{
		ApdexThresholdSeconds:  0.5,
		CollectAnalyticsEvents: true,
		CollectCustomEvents:    true,
		CollectTraces:          true,
		CollectErrors:          true,
		CollectErrorEvents:     true,
		// No transactions should be sampled before the application is
		// connected.
		AdaptiveSampler: SampleNothing{},
	}
}

// CalculateApdexThreshold calculates the apdex threshold.
func CalculateApdexThreshold(c *ConnectReply, txnName string) time.Duration {
	if t, ok := c.KeyTxnApdex[txnName]; ok {
		return floatSecondsToDuration(t)
	}
	return floatSecondsToDuration(c.ApdexThresholdSeconds)
}

// CreateFullTxnName uses collector rules and the appropriate metric prefix to
// construct the full transaction metric name from the name given by the
// consumer.
func CreateFullTxnName(input string, reply *ConnectReply, isWeb bool) string {
	var afterURLRules string
	if "" != input {
		afterURLRules = reply.URLRules.Apply(input)
		if "" == afterURLRules {
			return ""
		}
	}

	prefix := backgroundMetricPrefix
	if isWeb {
		prefix = webMetricPrefix
	}

	var beforeNameRules string
	if strings.HasPrefix(afterURLRules, "/") {
		beforeNameRules = prefix + afterURLRules
	} else {
		beforeNameRules = prefix + "/" + afterURLRules
	}

	afterNameRules := reply.TxnNameRules.Apply(beforeNameRules)
	if "" == afterNameRules {
		return ""
	}

	return reply.SegmentTerms.apply(afterNameRules)
}
