module github.com/imgproxy/imgproxy/v3

go 1.16

require (
	cloud.google.com/go/storage v1.22.1
	github.com/Azure/azure-storage-blob-go v0.15.0
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/airbrake/gobrake/v5 v5.5.1
	github.com/aws/aws-sdk-go v1.44.32
	github.com/benesch/cgosymbolizer v0.0.0-20190515212042-bec6fe6e597b
	github.com/bugsnag/bugsnag-go/v2 v2.1.2
	github.com/fsouza/fake-gcs-server v1.38.2
	github.com/getsentry/sentry-go v0.13.0
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/honeybadger-io/honeybadger-go v0.5.0
	github.com/ianlancetaylor/cgosymbolizer v0.0.0-20220405231054-a1ae3e4bba26 // indirect
	github.com/johannesboyne/gofakes3 v0.0.0-20220517215058-83a58ec253b6
	github.com/matoous/go-nanoid/v2 v2.0.0
	github.com/ncw/swift/v2 v2.0.1
	github.com/newrelic/go-agent/v3 v3.16.1
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/prometheus/client_golang v1.12.2
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.2
	github.com/tdewolff/parse/v2 v2.6.0
	github.com/trimmer-io/go-xmp v1.0.0
	go.uber.org/automaxprocs v1.5.1
	golang.org/x/image v0.0.0-20220601225756-64ec528b34cd
	golang.org/x/net v0.0.0-20220607020251-c690dde0001d
	golang.org/x/sys v0.0.0-20220610221304-9f5ed59c137d
	golang.org/x/text v0.3.7
	google.golang.org/api v0.83.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.38.1
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/shirou/gopsutil => github.com/shirou/gopsutil v2.20.9+incompatible

replace github.com/go-chi/chi/v4 => github.com/go-chi/chi v4.0.0+incompatible
