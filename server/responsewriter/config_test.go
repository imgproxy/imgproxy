package responsewriter

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type ResponseWriterConfigSuite struct {
	suite.Suite
}

func (s *ResponseWriterConfigSuite) SetupSuite() {
	logrus.SetOutput(io.Discard)
}

func (s *ResponseWriterConfigSuite) TearDownSuite() {
	logrus.SetOutput(os.Stdout)
}

func (s *ResponseWriterConfigSuite) TestLoadingVaryValueFromEnv() {
	defaultEnv := map[string]string{
		"IMGPROXY_AUTO_WEBP":           "",
		"IMGPROXY_ENFORCE_WEBP":        "",
		"IMGPROXY_AUTO_AVIF":           "",
		"IMGPROXY_ENFORCE_AVIF":        "",
		"IMGPROXY_AUTO_JXL":            "",
		"IMGPROXY_ENFORCE_JXL":         "",
		"IMGPROXY_ENABLE_CLIENT_HINTS": "",
	}

	testCases := []struct {
		name     string
		env      map[string]string
		expected string
	}{
		{
			name:     "AutoWebP",
			env:      map[string]string{"IMGPROXY_AUTO_WEBP": "true"},
			expected: "Accept",
		},
		{
			name:     "EnforceWebP",
			env:      map[string]string{"IMGPROXY_ENFORCE_WEBP": "true"},
			expected: "Accept",
		},
		{
			name:     "AutoAVIF",
			env:      map[string]string{"IMGPROXY_AUTO_AVIF": "true"},
			expected: "Accept",
		},
		{
			name:     "EnforceAVIF",
			env:      map[string]string{"IMGPROXY_ENFORCE_AVIF": "true"},
			expected: "Accept",
		},
		{
			name:     "AutoJXL",
			env:      map[string]string{"IMGPROXY_AUTO_JXL": "true"},
			expected: "Accept",
		},
		{
			name:     "EnforceJXL",
			env:      map[string]string{"IMGPROXY_ENFORCE_JXL": "true"},
			expected: "Accept",
		},
		{
			name:     "EnableClientHints",
			env:      map[string]string{"IMGPROXY_ENABLE_CLIENT_HINTS": "true"},
			expected: "Sec-CH-DPR, DPR, Sec-CH-Width, Width",
		},
		{
			name: "Combined",
			env: map[string]string{
				"IMGPROXY_AUTO_WEBP":           "true",
				"IMGPROXY_ENABLE_CLIENT_HINTS": "true",
			},
			expected: "Accept, Sec-CH-DPR, DPR, Sec-CH-Width, Width",
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%v", tc.env), func() {
			// Set default environment variables
			for key, value := range defaultEnv {
				s.T().Setenv(key, value)
			}
			// Set environment variables
			for key, value := range tc.env {
				s.T().Setenv(key, value)
			}

			// TODO: Remove when we removed global config
			config.Reset()
			config.Configure()

			// Load config
			cfg, err := LoadConfigFromEnv(nil)

			// Assert expected values
			s.Require().NoError(err)
			s.Require().Equal(tc.expected, cfg.VaryValue)
		})
	}
}

func TestResponseWriterConfig(t *testing.T) {
	suite.Run(t, new(ResponseWriterConfigSuite))
}
