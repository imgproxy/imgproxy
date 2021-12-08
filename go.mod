module github.com/imgproxy/imgproxy/v3

go 1.16

require (
	cloud.google.com/go/storage v1.18.2
	github.com/Azure/azure-storage-blob-go v0.14.0
	github.com/Microsoft/go-winio v0.5.1 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/airbrake/gobrake/v5 v5.1.1
	github.com/aws/aws-sdk-go v1.42.20
	github.com/benesch/cgosymbolizer v0.0.0-20190515212042-bec6fe6e597b
	github.com/bugsnag/bugsnag-go/v2 v2.1.2
	github.com/getsentry/sentry-go v0.11.0
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/honeybadger-io/honeybadger-go v0.5.0
	github.com/ianlancetaylor/cgosymbolizer v0.0.0-20210303021718-7cb7081f6b86 // indirect
	github.com/matoous/go-nanoid/v2 v2.0.0
	github.com/newrelic/go-agent/v3 v3.15.2
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/automaxprocs v1.4.0
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410
	golang.org/x/net v0.0.0-20211208012354-db4efeb81f4b
	golang.org/x/sys v0.0.0-20211205182925-97ca703d548d
	golang.org/x/text v0.3.7
	google.golang.org/api v0.61.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.34.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/shirou/gopsutil => github.com/shirou/gopsutil v2.20.9+incompatible
