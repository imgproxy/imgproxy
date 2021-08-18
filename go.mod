module github.com/imgproxy/imgproxy/v2

go 1.13

require (
	cloud.google.com/go/storage v1.15.0
	github.com/Azure/azure-storage-blob-go v0.13.0
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/airbrake/gobrake/v5 v5.0.3
	github.com/aws/aws-sdk-go v1.38.65
	github.com/benesch/cgosymbolizer v0.0.0-20190515212042-bec6fe6e597b
	github.com/bugsnag/bugsnag-go/v2 v2.1.1
	github.com/getsentry/sentry-go v0.11.0
	github.com/google/uuid v1.1.2
	github.com/honeybadger-io/honeybadger-go v0.5.0
	github.com/ianlancetaylor/cgosymbolizer v0.0.0-20210303021718-7cb7081f6b86 // indirect
	github.com/matoous/go-nanoid/v2 v2.0.0
	github.com/newrelic/go-agent/v3 v3.13.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/image v0.0.0-20201208152932-35266b937fa6
	golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420
	golang.org/x/sys v0.0.0-20210603125802-9665404d3644
	golang.org/x/text v0.3.6
	google.golang.org/api v0.48.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/shirou/gopsutil => github.com/shirou/gopsutil v2.20.9+incompatible
