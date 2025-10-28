package clientfeatures

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/logger"
)

type detectorTestCase struct {
	name     string
	config   Config
	header   map[string]string
	expected Features
}

type ClientFeaturesDetectorSuite struct {
	suite.Suite
}

func (s *ClientFeaturesDetectorSuite) SetupSuite() {
	logger.Mute()
}

func (s *ClientFeaturesDetectorSuite) TearDownSuite() {
	logger.Unmute()
}

func (s *ClientFeaturesDetectorSuite) runTestCases(testCases []detectorTestCase) {
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			detector := NewDetector(&tc.config)

			header := make(http.Header)
			for k, v := range tc.header {
				header.Set(k, v)
			}

			features := detector.Features(header)
			s.Require().Equal(tc.expected, features)
		})
	}
}

func (s *ClientFeaturesDetectorSuite) TestFeaturesAutoFormats() {
	s.runTestCases([]detectorTestCase{
		{
			name: "AutoWebP_ConainsWebP",
			config: Config{
				AutoWebp: true,
			},
			header: map[string]string{
				"Accept": "image/webp,image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{
				PreferWebP: true,
			},
		},
		{
			name: "AutoWebP_DoesNotContainWebP",
			config: Config{
				AutoWebp: true,
			},
			header: map[string]string{
				"Accept": "image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{},
		},
		{
			name: "EnforceWebP_ContainsWebP",
			config: Config{
				EnforceWebp: true,
			},
			header: map[string]string{
				"Accept": "image/webp,image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{
				PreferWebP:  true,
				EnforceWebP: true,
			},
		},
		{
			name: "EnforceWebP_DoesNotContainWebP",
			config: Config{
				EnforceWebp: true,
			},
			header: map[string]string{
				"Accept": "image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{},
		},
		{
			name: "AutoAvif_ContainsAvif",
			config: Config{
				AutoAvif: true,
			},
			header: map[string]string{
				"Accept": "image/avif,image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{
				PreferAvif: true,
			},
		},
		{
			name: "AutoAvif_DoesNotContainAvif",
			config: Config{
				AutoAvif: true,
			},
			header: map[string]string{
				"Accept": "image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{},
		},
		{
			name: "EnforceAvif_ContainsAvif",
			config: Config{
				EnforceAvif: true,
			},
			header: map[string]string{
				"Accept": "image/avif,image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{
				PreferAvif:  true,
				EnforceAvif: true,
			},
		},
		{
			name: "EnforceAvif_DoesNotContainAvif",
			config: Config{
				EnforceAvif: true,
			},
			header: map[string]string{
				"Accept": "image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{},
		},
		{
			name: "AutoJXL_ContainsJXL",
			config: Config{
				AutoJxl: true,
			},
			header: map[string]string{
				"Accept": "image/jxl,image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{
				PreferJxl: true,
			},
		},
		{
			name: "AutoJXL_DoesNotContainJXL",
			config: Config{
				AutoJxl: true,
			},
			header: map[string]string{
				"Accept": "image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{},
		},
		{
			name: "EnforceJXL_ContainsJXL",
			config: Config{
				EnforceJxl: true,
			},
			header: map[string]string{
				"Accept": "image/jxl,image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{
				PreferJxl:  true,
				EnforceJxl: true,
			},
		},
		{
			name: "EnforceJXL_DoesNotContainJXL",
			config: Config{
				EnforceJxl: true,
			},
			header: map[string]string{
				"Accept": "image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{},
		},
		{
			name: "NoneEnabled_ContainsAll",
			config: Config{
				AutoWebp:    false,
				EnforceWebp: false,
				AutoAvif:    false,
				EnforceAvif: false,
				AutoJxl:     false,
				EnforceJxl:  false,
			},
			header: map[string]string{
				"Accept": "image/webp,image/avif,image/jxl,image/apng,image/*,*/*;q=0.8",
			},
			expected: Features{},
		},
	})
}

func (s *ClientFeaturesDetectorSuite) TestFeaturesClientHintsDPR() {
	s.runTestCases([]detectorTestCase{
		{
			name: "ClientHintsEnabled_ValidDPR",
			config: Config{
				EnableClientHints: true,
			},
			header: map[string]string{
				"DPR": "1.5",
			},
			expected: Features{
				ClientHintsDPR: 1.5,
			},
		},
		{
			name: "ClientHintsEnabled_ValidSecChDPR",
			config: Config{
				EnableClientHints: true,
			},
			header: map[string]string{
				"Sec-CH-DPR": "2.0",
			},
			expected: Features{
				ClientHintsDPR: 2.0,
			},
		},
		{
			name: "ClientHintsEnabled_ValidDprAndSecChDPR",
			config: Config{
				EnableClientHints: true,
			},
			header: map[string]string{
				"DPR":        "3.0",
				"Sec-CH-DPR": "2.5",
			},
			expected: Features{
				ClientHintsDPR: 2.5,
			},
		},
		{
			name: "ClientHintsEnabled_InvalidDPR_Negative",
			config: Config{
				EnableClientHints: true,
			},
			header: map[string]string{
				"DPR": "-1.0",
			},
			expected: Features{},
		},
		{
			name: "ClientHintsEnabled_InvalidDPR_TooHigh",
			config: Config{
				EnableClientHints: true,
			},
			header: map[string]string{
				"DPR": "10.0",
			},
			expected: Features{},
		},
		{
			name: "ClientHintsEnabled_InvalidDPR_NonNumeric",
			config: Config{
				EnableClientHints: true,
			},
			header: map[string]string{
				"DPR": "abc",
			},
			expected: Features{},
		},
		{
			name: "ClientHintsDisabled",
			config: Config{
				EnableClientHints: false,
			},
			header: map[string]string{
				"DPR":        "2.0",
				"Sec-CH-DPR": "3.0",
			},
			expected: Features{},
		},
	})
}

func (s *ClientFeaturesDetectorSuite) TestFeaturesClientHintsWidth() {
	s.runTestCases([]detectorTestCase{
		{
			name: "ClientHintsEnabled_ValidWidth",
			config: Config{
				EnableClientHints: true,
			},
			header: map[string]string{
				"Width": "800",
			},
			expected: Features{
				ClientHintsWidth: 800,
			},
		},
		{
			name: "ClientHintsEnabled_ValidSecChWidth",
			config: Config{
				EnableClientHints: true,
			},
			header: map[string]string{
				"Sec-CH-Width": "1024",
			},
			expected: Features{
				ClientHintsWidth: 1024,
			},
		},
		{
			name: "ClientHintsEnabled_ValidWidthAndSecChWidth",
			config: Config{
				EnableClientHints: true,
			},
			header: map[string]string{
				"Width":        "1280",
				"Sec-CH-Width": "1440",
			},
			expected: Features{
				ClientHintsWidth: 1440,
			},
		},
		{
			name: "ClientHintsEnabled_InvalidWidth_Negative",
			config: Config{
				EnableClientHints: true,
			},
			header: map[string]string{
				"Width": "-800",
			},
			expected: Features{},
		},
		{
			name: "ClientHintsEnabled_InvalidWidth_NonNumeric",
			config: Config{
				EnableClientHints: true,
			},
			header: map[string]string{
				"Width": "abc",
			},
			expected: Features{},
		},
		{
			name: "ClientHintsDisabled",
			config: Config{
				EnableClientHints: false,
			},
			header: map[string]string{
				"Width":        "800",
				"Sec-CH-Width": "1024",
			},
			expected: Features{},
		},
	})
}

func (s *ClientFeaturesDetectorSuite) TestSetVary() {
	testCases := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name: "AutoWebP_Enabled",
			config: Config{
				AutoWebp: true,
			},
			expected: "Accept",
		},
		{
			name: "EnforceWebP_Enabled",
			config: Config{
				EnforceWebp: true,
			},
			expected: "Accept",
		},
		{
			name: "AutoAvif_Enabled",
			config: Config{
				AutoAvif: true,
			},
			expected: "Accept",
		},
		{
			name: "EnforceAvif_Enabled",
			config: Config{
				EnforceAvif: true,
			},
			expected: "Accept",
		},
		{
			name: "AutoJXL_Enabled",
			config: Config{
				AutoJxl: true,
			},
			expected: "Accept",
		},
		{
			name: "EnforceJXL_Enabled",
			config: Config{
				EnforceJxl: true,
			},
			expected: "Accept",
		},
		{
			name: "EnableClientHints_Enabled",
			config: Config{
				EnableClientHints: true,
			},
			expected: "Sec-Ch-Dpr, Dpr, Sec-Ch-Width, Width",
		},
		{
			name: "Combined",
			config: Config{
				AutoWebp:          true,
				EnableClientHints: true,
			},
			expected: "Accept, Sec-Ch-Dpr, Dpr, Sec-Ch-Width, Width",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			detector := NewDetector(&tc.config)
			header := http.Header{}
			detector.SetVary(header)
			s.Require().Equal(tc.expected, header.Get(httpheaders.Vary))
		})
	}
}

func TestClientFeaturesDetector(t *testing.T) {
	suite.Run(t, new(ClientFeaturesDetectorSuite))
}
