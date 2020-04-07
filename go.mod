module github.com/imgproxy/imgproxy/v2

go 1.11

require (
	cloud.google.com/go/storage v1.6.0
	github.com/StackExchange/wmi v0.0.0-20180116203802-5d049714c4a6 // indirect
	github.com/aws/aws-sdk-go v1.30.4
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/bugsnag/bugsnag-go v1.5.3
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/getsentry/sentry-go v0.5.1
	github.com/go-ole/go-ole v1.2.2 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/google/uuid v1.1.0 // indirect
	github.com/honeybadger-io/honeybadger-go v0.5.0
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1 // indirect
	github.com/matoous/go-nanoid v1.1.1-0.20200226125206-b0a1054fe39d
	github.com/newrelic/go-agent v2.16.3+incompatible
	github.com/prometheus/client_golang v0.9.4
	github.com/sirupsen/logrus v1.5.0
	github.com/stretchr/testify v1.5.1
	golang.org/x/image v0.0.0-20190802002840-cff245a6509b
	golang.org/x/net v0.0.0-20200222125558-5a598a2470a0
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae
	google.golang.org/api v0.21.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
