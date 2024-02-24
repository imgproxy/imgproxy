package otel

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
)

type OtelTestSuite struct{ suite.Suite }

func (s *OtelTestSuite) SetupSuite() {
	logrus.SetOutput(io.Discard)
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

	require.True(s.T(), config.OpenTelemetryEnable)
	require.Equal(s.T(), "https://otel_endpoint:1234", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	require.Equal(s.T(), "", os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigEndpointGrpcProtocol() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_ENDPOINT", "otel_endpoint:1234")
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_PROTOCOL", "grpc")

	mapDeprecatedConfig()

	require.True(s.T(), config.OpenTelemetryEnable)
	require.Equal(s.T(), "https://otel_endpoint:1234", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	require.Equal(s.T(), "grpc", os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigEndpointGrpcProtocolInsecure() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_ENDPOINT", "otel_endpoint:1234")
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_PROTOCOL", "grpc")
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_GRPC_INSECURE", "1")

	mapDeprecatedConfig()

	require.True(s.T(), config.OpenTelemetryEnable)
	require.Equal(s.T(), "http://otel_endpoint:1234", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	require.Equal(s.T(), "grpc", os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigEndpointHttpsProtocol() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_ENDPOINT", "otel_endpoint:1234")
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_PROTOCOL", "https")

	mapDeprecatedConfig()

	require.True(s.T(), config.OpenTelemetryEnable)
	require.Equal(s.T(), "https://otel_endpoint:1234", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	require.Equal(s.T(), "https", os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigEndpointHttpProtocol() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_ENDPOINT", "otel_endpoint:1234")
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_PROTOCOL", "http")

	mapDeprecatedConfig()

	require.True(s.T(), config.OpenTelemetryEnable)
	require.Equal(s.T(), "http://otel_endpoint:1234", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	require.Equal(s.T(), "http", os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigServiceName() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_SERVICE_NAME", "testtest")

	config.OpenTelemetryEnable = true
	mapDeprecatedConfig()

	require.Equal(s.T(), "testtest", os.Getenv("OTEL_SERVICE_NAME"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigPropagators() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_PROPAGATORS", "testtest")

	config.OpenTelemetryEnable = true
	mapDeprecatedConfig()

	require.Equal(s.T(), "testtest", os.Getenv("OTEL_PROPAGATORS"))
}

func (s *OtelTestSuite) TestMapDeprecatedConfigConnectionTimeout() {
	os.Setenv("IMGPROXY_OPEN_TELEMETRY_CONNECTION_TIMEOUT", "15")

	config.OpenTelemetryEnable = true
	mapDeprecatedConfig()

	require.Equal(s.T(), "15000", os.Getenv("OTEL_EXPORTER_OTLP_TIMEOUT"))
}

func TestPresets(t *testing.T) {
	suite.Run(t, new(OtelTestSuite))
}
