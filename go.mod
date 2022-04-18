module github.com/imgproxy/imgproxy/v3

go 1.16

require (
	cloud.google.com/go/storage v1.22.0
	github.com/Azure/azure-storage-blob-go v0.14.0
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/airbrake/gobrake/v5 v5.3.0
	github.com/aws/aws-sdk-go v1.43.41
	github.com/benesch/cgosymbolizer v0.0.0-20190515212042-bec6fe6e597b
	github.com/bugsnag/bugsnag-go/v2 v2.1.2
	github.com/fsouza/fake-gcs-server v1.37.10-0.20220407170115-d717316ed8a8
	github.com/getsentry/sentry-go v0.13.0
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/honeybadger-io/honeybadger-go v0.5.0
	github.com/ianlancetaylor/cgosymbolizer v0.0.0-20220405231054-a1ae3e4bba26 // indirect
	github.com/johannesboyne/gofakes3 v0.0.0-20220314170512-33c13122505e
	github.com/matoous/go-nanoid/v2 v2.0.0
	github.com/ncw/swift/v2 v2.0.1
	github.com/newrelic/go-agent/v3 v3.15.2
	github.com/prometheus/client_golang v1.12.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1
	go.uber.org/automaxprocs v1.5.0
	golang.org/x/image v0.0.0-20220321031419-a8550c1d254a
	golang.org/x/net v0.0.0-20220403103023-749bd193bc2b
	golang.org/x/sys v0.0.0-20220405210540-1e041c57c461
	golang.org/x/text v0.3.7
	google.golang.org/api v0.74.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.37.1
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/shirou/gopsutil => github.com/shirou/gopsutil v2.20.9+incompatible

replace github.com/go-chi/chi/v4 => github.com/go-chi/chi v4.0.0+incompatible
