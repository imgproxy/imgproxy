package cloudwatch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
	"github.com/sirupsen/logrus"
)

type GaugeFunc func() float64

type gauge struct {
	unit string
	f    GaugeFunc
}

type bufferStats struct {
	count         int
	sum, min, max int
}

var (
	enabled bool

	client *cloudwatch.CloudWatch

	gauges      = make(map[string]gauge)
	gaugesMutex sync.RWMutex

	collectorCtx       context.Context
	collectorCtxCancel context.CancelFunc

	bufferDefaultSizes = make(map[string]int)
	bufferMaxSizes     = make(map[string]int)
	bufferSizeStats    = make(map[string]*bufferStats)
	bufferStatsMutex   sync.Mutex
)

func Init() error {
	if len(config.CloudWatchServiceName) == 0 {
		return nil
	}

	conf := aws.NewConfig()

	if len(config.CloudWatchRegion) > 0 {
		conf = conf.WithRegion(config.CloudWatchRegion)
	}

	sess, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("Can't create CloudWatch session: %s", err)
	}

	if sess.Config.Region == nil || len(*sess.Config.Region) == 0 {
		sess.Config.Region = aws.String("us-west-1")
	}

	client = cloudwatch.New(sess, conf)

	collectorCtx, collectorCtxCancel = context.WithCancel(context.Background())

	go runMetricsCollector()

	enabled = true

	return nil
}

func Stop() {
	if enabled {
		collectorCtxCancel()
	}
}

func Enabled() bool {
	return enabled
}

func AddGaugeFunc(name, unit string, f GaugeFunc) {
	gaugesMutex.Lock()
	defer gaugesMutex.Unlock()

	gauges[name] = gauge{unit: unit, f: f}
}

func ObserveBufferSize(t string, size int) {
	if enabled {
		bufferStatsMutex.Lock()
		defer bufferStatsMutex.Unlock()

		sizef := size

		stats, ok := bufferSizeStats[t]
		if !ok {
			stats = &bufferStats{count: 1, sum: sizef, min: sizef, max: sizef}
			bufferSizeStats[t] = stats
			return
		}

		stats.count += 1
		stats.sum += sizef
		stats.min = imath.Min(stats.min, sizef)
		stats.max = imath.Max(stats.max, sizef)
	}
}

func SetBufferDefaultSize(t string, size int) {
	if enabled {
		bufferStatsMutex.Lock()
		defer bufferStatsMutex.Unlock()

		bufferDefaultSizes[t] = size
	}
}

func SetBufferMaxSize(t string, size int) {
	if enabled {
		bufferStatsMutex.Lock()
		defer bufferStatsMutex.Unlock()

		bufferMaxSizes[t] = size
	}
}

func runMetricsCollector() {
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()

	dimension := &cloudwatch.Dimension{
		Name:  aws.String("ServiceName"),
		Value: aws.String(config.CloudWatchServiceName),
	}

	bufferDimensions := make(map[string]*cloudwatch.Dimension)
	bufferDimension := func(t string) *cloudwatch.Dimension {
		if d, ok := bufferDimensions[t]; ok {
			return d
		}

		d := &cloudwatch.Dimension{
			Name:  aws.String("BufferType"),
			Value: aws.String(t),
		}

		bufferDimensions[t] = d

		return d
	}

	for {
		select {
		case <-tick.C:
			metricsCount := len(gauges) + len(bufferDefaultSizes) + len(bufferMaxSizes) + len(bufferSizeStats) + 3
			metrics := make([]*cloudwatch.MetricDatum, 0, metricsCount)

			func() {
				gaugesMutex.RLock()
				defer gaugesMutex.RUnlock()

				for name, g := range gauges {
					metrics = append(metrics, &cloudwatch.MetricDatum{
						Dimensions: []*cloudwatch.Dimension{dimension},
						MetricName: aws.String(name),
						Unit:       aws.String(g.unit),
						Value:      aws.Float64(g.f()),
					})
				}
			}()

			func() {
				bufferStatsMutex.Lock()
				defer bufferStatsMutex.Unlock()

				for t, size := range bufferDefaultSizes {
					metrics = append(metrics, &cloudwatch.MetricDatum{
						Dimensions: []*cloudwatch.Dimension{dimension, bufferDimension(t)},
						MetricName: aws.String("BufferDefaultSize"),
						Unit:       aws.String("Bytes"),
						Value:      aws.Float64(float64(size)),
					})
				}

				for t, size := range bufferMaxSizes {
					metrics = append(metrics, &cloudwatch.MetricDatum{
						Dimensions: []*cloudwatch.Dimension{dimension, bufferDimension(t)},
						MetricName: aws.String("BufferMaximumSize"),
						Unit:       aws.String("Bytes"),
						Value:      aws.Float64(float64(size)),
					})
				}

				for t, stats := range bufferSizeStats {
					metrics = append(metrics, &cloudwatch.MetricDatum{
						Dimensions: []*cloudwatch.Dimension{dimension, bufferDimension(t)},
						MetricName: aws.String("BufferSize"),
						Unit:       aws.String("Bytes"),
						StatisticValues: &cloudwatch.StatisticSet{
							SampleCount: aws.Float64(float64(stats.count)),
							Sum:         aws.Float64(float64(stats.sum)),
							Minimum:     aws.Float64(float64(stats.min)),
							Maximum:     aws.Float64(float64(stats.max)),
						},
					})
				}
			}()

			metrics = append(metrics, &cloudwatch.MetricDatum{
				Dimensions: []*cloudwatch.Dimension{dimension},
				MetricName: aws.String("RequestsInProgress"),
				Unit:       aws.String("Count"),
				Value:      aws.Float64(stats.RequestsInProgress()),
			})

			metrics = append(metrics, &cloudwatch.MetricDatum{
				Dimensions: []*cloudwatch.Dimension{dimension},
				MetricName: aws.String("ImagesInProgress"),
				Unit:       aws.String("Count"),
				Value:      aws.Float64(stats.ImagesInProgress()),
			})

			metrics = append(metrics, &cloudwatch.MetricDatum{
				Dimensions: []*cloudwatch.Dimension{dimension},
				MetricName: aws.String("ConcurrencyUtilization"),
				Unit:       aws.String("Percent"),
				Value: aws.Float64(
					stats.RequestsInProgress() / float64(config.Concurrency) * 100.0,
				),
			})

			_, err := client.PutMetricData(&cloudwatch.PutMetricDataInput{
				Namespace:  aws.String(config.CloudWatchNamespace),
				MetricData: metrics,
			})
			if err != nil {
				logrus.Warnf("Can't send CloudWatch metrics: %s", err)
			}
		case <-collectorCtx.Done():
			return
		}
	}
}
