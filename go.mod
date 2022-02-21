module github.com/imgproxy/imgproxy/v3

go 1.16

require (
	cloud.google.com/go/storage v1.21.0
	github.com/Azure/azure-storage-blob-go v0.14.0
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/airbrake/gobrake/v5 v5.3.0
	github.com/aws/aws-sdk-go v1.43.2
	github.com/benesch/cgosymbolizer v0.0.0-20190515212042-bec6fe6e597b
	github.com/bugsnag/bugsnag-go/v2 v2.1.2
	github.com/getsentry/sentry-go v0.12.0
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/honeybadger-io/honeybadger-go v0.5.0
	github.com/ianlancetaylor/cgosymbolizer v0.0.0-20220217162856-c813f11194b9 // indirect
	github.com/matoous/go-nanoid/v2 v2.0.0
	github.com/newrelic/go-agent/v3 v3.15.2
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/prometheus/client_golang v1.12.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/automaxprocs v1.4.0
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158
	golang.org/x/text v0.3.7
	google.golang.org/api v0.69.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.36.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/shirou/gopsutil => github.com/shirou/gopsutil v2.20.9+incompatible
