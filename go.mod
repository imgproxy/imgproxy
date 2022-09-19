module github.com/imgproxy/imgproxy/v3

go 1.16

require (
	cloud.google.com/go/storage v1.25.0
	github.com/Azure/azure-storage-blob-go v0.15.0
	github.com/DataDog/datadog-go/v5 v5.1.1
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/airbrake/gobrake/v5 v5.5.2
	github.com/aws/aws-sdk-go v1.44.81
	github.com/benesch/cgosymbolizer v0.0.0-20190515212042-bec6fe6e597b
	github.com/bugsnag/bugsnag-go/v2 v2.1.2
	github.com/fsouza/fake-gcs-server v1.38.4
	github.com/getsentry/sentry-go v0.13.0
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/honeybadger-io/honeybadger-go v0.5.0
	github.com/ianlancetaylor/cgosymbolizer v0.0.0-20220405231054-a1ae3e4bba26 // indirect
	github.com/johannesboyne/gofakes3 v0.0.0-20220627085814-c3ac35da23b2
	github.com/matoous/go-nanoid/v2 v2.0.0
	github.com/ncw/swift/v2 v2.0.1
	github.com/newrelic/go-agent/v3 v3.18.1
	github.com/newrelic/newrelic-telemetry-sdk-go v0.8.1
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/prometheus/client_golang v1.13.0
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.0
	github.com/tdewolff/parse/v2 v2.6.2
	github.com/trimmer-io/go-xmp v1.0.0
	go.uber.org/automaxprocs v1.5.1
	golang.org/x/image v0.0.0-20220722155232-062f8c9fd539
	golang.org/x/net v0.0.0-20220909164309-bea034e7d591
	golang.org/x/sys v0.0.0-20220818161305-2296e01440c6
	google.golang.org/api v0.96.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.41.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/shirou/gopsutil => github.com/shirou/gopsutil v2.20.9+incompatible

replace github.com/go-chi/chi/v4 => github.com/go-chi/chi v4.0.0+incompatible
