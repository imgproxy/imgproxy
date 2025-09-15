package otel

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/logger"
)

type OtelTestSuite struct{ suite.Suite }

func (s *OtelTestSuite) SetupSuite() {
	logger.Mute()
}

func (s *OtelTestSuite) TearDownSuite() {
	logger.Unmute()
}

func (s *OtelTestSuite) SetupTest() {
	for _, env := range os.Environ() {
		keyVal := strings.Split(env, "=")
		if strings.HasPrefix(keyVal[0], "OTEL_") || strings.HasPrefix(keyVal[0], "IMGPROXY_OPEN_TELEMETRY_") {
			os.Unsetenv(keyVal[0])
		}
	}

	config.Reset()
}

func (s *OtelTestSuite) TestMapDeprecatedConfigEndpointNoProtocol() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_ENDPOINT", "otel_endpoint:1234")

	mapDeprecatedConfig()

	s.Require().True(config.OpenTelemetryEnable)
	s.Require().Equal("https://otel_endpoint:1234", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	s.Require().Empty(os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigEndpointGrpcProtocol() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_ENDPOINT", "otel_endpoint:1234")
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_PROTOCOL", "grpc")

	mapDeprecatedConfig()

	s.Require().True(config.OpenTelemetryEnable)
	s.Require().Equal("https://otel_endpoint:1234", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	s.Require().Equal("grpc", os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigEndpointGrpcProtocolInsecure() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_ENDPOINT", "otel_endpoint:1234")
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_PROTOCOL", "grpc")
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_GRPC_INSECURE", "1")

	mapDeprecatedConfig()

	s.Require().True(config.OpenTelemetryEnable)
	s.Require().Equal("http://otel_endpoint:1234", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	s.Require().Equal("grpc", os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigEndpointHttpsProtocol() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_ENDPOINT", "otel_endpoint:1234")
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_PROTOCOL", "https")

	mapDeprecatedConfig()

	s.Require().True(config.OpenTelemetryEnable)
	s.Require().Equal("https://otel_endpoint:1234", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	s.Require().Equal("https", os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigEndpointHttpProtocol() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_ENDPOINT", "otel_endpoint:1234")
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_PROTOCOL", "http")

	mapDeprecatedConfig()

	s.Require().True(config.OpenTelemetryEnable)
	s.Require().Equal("http://otel_endpoint:1234", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	s.Require().Equal("http", os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigServiceName() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_SERVICE_NAME", "testtest")

	config.OpenTelemetryEnable = true
	mapDeprecatedConfig()

	s.Require().Equal("testtest", os.Getenv("OTEL_SERVICE_NAME"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigPropagators() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_PROPAGATORS", "testtest")

	config.OpenTelemetryEnable = true
	mapDeprecatedConfig()

	s.Require().Equal("testtest", os.Getenv("OTEL_PROPAGATORS"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigConnectionTimeout() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_CONNECTION_TIMEOUT", "15")

	config.OpenTelemetryEnable = true
	mapDeprecatedConfig()

	s.Require().Equal("15000", os.Getenv("OTEL_EXPORTER_OTLP_TIMEOUT"))
}

func TestPresets(t *testing.T) {
	suite.Run(t, new(OtelTestSuite))
}
