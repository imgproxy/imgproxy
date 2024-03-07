package cloudwatch

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cloudwatchTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
)

type GaugeFunc func() float64

type gauge struct {
	unit cloudwatchTypes.StandardUnit
	f    GaugeFunc
}

type bufferStats struct {
	count         int
	sum, min, max int
}

var (
	enabled bool

	client *cloudwatch.Client

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

	conf, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return fmt.Errorf("can't load CloudWatch config: %s", err)
	}

	if len(config.CloudWatchRegion) != 0 {
		conf.Region = config.CloudWatchRegion
	}

	if len(conf.Region) == 0 {
		conf.Region = "us-west-1"
	}

	client = cloudwatch.NewFromConfig(conf)

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

	standardUnit := cloudwatchTypes.StandardUnit(unit)

	if !slices.Contains(cloudwatchTypes.StandardUnitNone.Values(), standardUnit) {
		panic(fmt.Errorf("Unknown CloudWatch unit: %s", unit))
	}

	gauges[name] = gauge{unit: standardUnit, f: f}
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

	dimension := cloudwatchTypes.Dimension{
		Name:  aws.String("ServiceName"),
		Value: aws.String(config.CloudWatchServiceName),
	}

	bufferDimensions := make(map[string]cloudwatchTypes.Dimension)
	bufferDimension := func(t string) cloudwatchTypes.Dimension {
		if d, ok := bufferDimensions[t]; ok {
			return d
		}

		d := cloudwatchTypes.Dimension{
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
			metrics := make([]cloudwatchTypes.MetricDatum, 0, metricsCount)

			func() {
				gaugesMutex.RLock()
				defer gaugesMutex.RUnlock()

				for name, g := range gauges {
					metrics = append(metrics, cloudwatchTypes.MetricDatum{
						Dimensions: []cloudwatchTypes.Dimension{dimension},
						MetricName: aws.String(name),
						Unit:       g.unit,
						Value:      aws.Float64(g.f()),
					})
				}
			}()

			func() {
				bufferStatsMutex.Lock()
				defer bufferStatsMutex.Unlock()

				for t, size := range bufferDefaultSizes {
					metrics = append(metrics, cloudwatchTypes.MetricDatum{
						Dimensions: []cloudwatchTypes.Dimension{dimension, bufferDimension(t)},
						MetricName: aws.String("BufferDefaultSize"),
						Unit:       cloudwatchTypes.StandardUnitBytes,
						Value:      aws.Float64(float64(size)),
					})
				}

				for t, size := range bufferMaxSizes {
					metrics = append(metrics, cloudwatchTypes.MetricDatum{
						Dimensions: []cloudwatchTypes.Dimension{dimension, bufferDimension(t)},
						MetricName: aws.String("BufferMaximumSize"),
						Unit:       cloudwatchTypes.StandardUnitBytes,
						Value:      aws.Float64(float64(size)),
					})
				}

				for t, stats := range bufferSizeStats {
					metrics = append(metrics, cloudwatchTypes.MetricDatum{
						Dimensions: []cloudwatchTypes.Dimension{dimension, bufferDimension(t)},
						MetricName: aws.String("BufferSize"),
						Unit:       cloudwatchTypes.StandardUnitBytes,
						StatisticValues: &cloudwatchTypes.StatisticSet{
							SampleCount: aws.Float64(float64(stats.count)),
							Sum:         aws.Float64(float64(stats.sum)),
							Minimum:     aws.Float64(float64(stats.min)),
							Maximum:     aws.Float64(float64(stats.max)),
						},
					})
				}
			}()

			metrics = append(metrics, cloudwatchTypes.MetricDatum{
				Dimensions: []cloudwatchTypes.Dimension{dimension},
				MetricName: aws.String("RequestsInProgress"),
				Unit:       cloudwatchTypes.StandardUnitCount,
				Value:      aws.Float64(stats.RequestsInProgress()),
			})

			metrics = append(metrics, cloudwatchTypes.MetricDatum{
				Dimensions: []cloudwatchTypes.Dimension{dimension},
				MetricName: aws.String("ImagesInProgress"),
				Unit:       cloudwatchTypes.StandardUnitCount,
				Value:      aws.Float64(stats.ImagesInProgress()),
			})

			metrics = append(metrics, cloudwatchTypes.MetricDatum{
				Dimensions: []cloudwatchTypes.Dimension{dimension},
				MetricName: aws.String("ConcurrencyUtilization"),
				Unit:       cloudwatchTypes.StandardUnitPercent,
				Value: aws.Float64(
					stats.RequestsInProgress() / float64(config.Workers) * 100.0,
				),
			})

			metrics = append(metrics, cloudwatchTypes.MetricDatum{
				Dimensions: []cloudwatchTypes.Dimension{dimension},
				MetricName: aws.String("WorkersUtilization"),
				Unit:       cloudwatchTypes.StandardUnitPercent,
				Value: aws.Float64(
					stats.RequestsInProgress() / float64(config.Workers) * 100.0,
				),
			})

			input := cloudwatch.PutMetricDataInput{
				Namespace:  aws.String(config.CloudWatchNamespace),
				MetricData: metrics,
			}

			func() {
				ctx, cancel := context.WithTimeout(collectorCtx, 30*time.Second)
				defer cancel()

				if _, err := client.PutMetricData(ctx, &input); err != nil {
					logrus.Warnf("Can't send CloudWatch metrics: %s", err)
				}
			}()
		case <-collectorCtx.Done():
			return
		}
	}
}
