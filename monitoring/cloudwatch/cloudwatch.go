package cloudwatch

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cloudwatchTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	vipsstats "github.com/imgproxy/imgproxy/v3/vips/stats"
)

const (
	// AWS CloudWatch PutMetrics timeout
	putMetricsTimeout = 30 * time.Second

	// default AWS region to set if neither aws env region nor config region are set
	defaultAwsRegion = "us-west-1"
)

// CloudWatch holds CloudWatch client and configuration
type CloudWatch struct {
	config *Config
	stats  *stats.Stats

	client *cloudwatch.Client

	collectorCtx       context.Context
	collectorCtxCancel context.CancelFunc
}

// New creates a new CloudWatch instance
func New(ctx context.Context, config *Config, stats *stats.Stats) (*CloudWatch, error) {
	if !config.Enabled() {
		return nil, nil
	}

	cw := &CloudWatch{
		config: config,
		stats:  stats,
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	conf, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't load CloudWatch config: %s", err)
	}

	if len(config.Region) > 0 {
		conf.Region = config.Region
	}

	if len(conf.Region) == 0 {
		conf.Region = defaultAwsRegion
	}

	cw.client = cloudwatch.NewFromConfig(conf)
	cw.collectorCtx, cw.collectorCtxCancel = context.WithCancel(ctx)

	go cw.runMetricsCollector()

	return cw, nil
}

// Stop stops the CloudWatch metrics collection
func (cw *CloudWatch) Stop(ctx context.Context) {
	if cw.collectorCtxCancel != nil {
		cw.collectorCtxCancel()
	}
}

// runMetricsCollector collects and sends metrics to CloudWatch
func (cw *CloudWatch) runMetricsCollector() {
	tick := time.NewTicker(cw.config.MetricsInterval)
	defer tick.Stop()

	dimension := cloudwatchTypes.Dimension{
		Name:  aws.String("ServiceName"),
		Value: aws.String(cw.config.ServiceName),
	}

	dimensions := []cloudwatchTypes.Dimension{dimension}

	namespace := aws.String(cw.config.Namespace)

	// metric names
	metricNameWorkers := aws.String("Workers")
	metricNameRequestsInProgress := aws.String("RequestsInProgress")
	metricNameImagesInProgress := aws.String("ImagesInProgress")
	metricNameConcurrencyUtilization := aws.String("ConcurrencyUtilization")
	metricNameWorkersUtilization := aws.String("WorkersUtilization")
	metricNameVipsMemory := aws.String("VipsMemory")
	metricNameVipsMaxMemory := aws.String("VipsMaxMemory")
	metricNameVipsAllocs := aws.String("VipsAllocs")

	for {
		select {
		case <-tick.C:
			// 8 is the number of metrics we send
			metrics := make([]cloudwatchTypes.MetricDatum, 0, 8)

			metrics = append(metrics, cloudwatchTypes.MetricDatum{
				Dimensions: dimensions,
				MetricName: metricNameWorkers,
				Unit:       cloudwatchTypes.StandardUnitCount,
				Value:      aws.Float64(float64(cw.stats.WorkersNumber)),
			})

			metrics = append(metrics, cloudwatchTypes.MetricDatum{
				Dimensions: dimensions,
				MetricName: metricNameRequestsInProgress,
				Unit:       cloudwatchTypes.StandardUnitCount,
				Value:      aws.Float64(cw.stats.RequestsInProgress()),
			})

			metrics = append(metrics, cloudwatchTypes.MetricDatum{
				Dimensions: dimensions,
				MetricName: metricNameImagesInProgress,
				Unit:       cloudwatchTypes.StandardUnitCount,
				Value:      aws.Float64(cw.stats.ImagesInProgress()),
			})

			metrics = append(metrics, cloudwatchTypes.MetricDatum{
				Dimensions: dimensions,
				MetricName: metricNameConcurrencyUtilization,
				Unit:       cloudwatchTypes.StandardUnitPercent,
				Value: aws.Float64(
					cw.stats.WorkersUtilization(),
				),
			})

			metrics = append(metrics, cloudwatchTypes.MetricDatum{
				Dimensions: dimensions,
				MetricName: metricNameWorkersUtilization,
				Unit:       cloudwatchTypes.StandardUnitPercent,
				Value: aws.Float64(
					cw.stats.WorkersUtilization(),
				),
			})

			metrics = append(metrics, cloudwatchTypes.MetricDatum{
				Dimensions: dimensions,
				MetricName: metricNameVipsMemory,
				Unit:       cloudwatchTypes.StandardUnitBytes,
				Value:      aws.Float64(vipsstats.Memory()),
			})

			metrics = append(metrics, cloudwatchTypes.MetricDatum{
				Dimensions: dimensions,
				MetricName: metricNameVipsMaxMemory,
				Unit:       cloudwatchTypes.StandardUnitBytes,
				Value:      aws.Float64(vipsstats.MemoryHighwater()),
			})

			metrics = append(metrics, cloudwatchTypes.MetricDatum{
				Dimensions: dimensions,
				MetricName: metricNameVipsAllocs,
				Unit:       cloudwatchTypes.StandardUnitCount,
				Value:      aws.Float64(vipsstats.Allocs()),
			})

			input := cloudwatch.PutMetricDataInput{
				Namespace:  namespace,
				MetricData: metrics,
			}

			func() {
				ctx, cancel := context.WithTimeout(cw.collectorCtx, putMetricsTimeout)
				defer cancel()

				if _, err := cw.client.PutMetricData(ctx, &input); err != nil {
					slog.Warn(fmt.Sprintf("can't send CloudWatch metrics: %s", err))
				}
			}()
		case <-cw.collectorCtx.Done():
			return
		}
	}
}

// StartRequest starts a new request span
func (cw *CloudWatch) StartRequest(
	ctx context.Context,
	rw http.ResponseWriter,
	r *http.Request,
) (context.Context, context.CancelFunc, http.ResponseWriter) {
	// CloudWatch does not support request tracing
	return ctx, func() {}, rw
}

// StartSpan starts a new span
func (cw *CloudWatch) StartSpan(
	ctx context.Context,
	name string,
	meta map[string]any,
) (context.Context, context.CancelFunc) {
	// CloudWatch does not support request tracing
	return ctx, func() {}
}

// SetMetadata sets metadata for the current span
func (cw *CloudWatch) SetMetadata(ctx context.Context, key string, value any) {
	// CloudWatch does not support request tracing
}

// SetError records an error in the current span
func (cw *CloudWatch) SendError(ctx context.Context, errType string, err error) {
	// CloudWatch does not support request tracing
}

// InjectHeaders adds monitoring headers to the provided HTTP headers.
func (cw *CloudWatch) InjectHeaders(ctx context.Context, headers http.Header) {
	// CloudWatch does not support request tracing
}
