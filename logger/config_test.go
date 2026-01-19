package logger

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyslogConfigLevel(t *testing.T) {
	tests := []struct {
		name              string
		loggerLvl         string
		syslogLvl         string
		expectedLoggerLvl slog.Leveler
		expectedSyslogLvl slog.Leveler
	}{
		{
			name:              "Defaults",
			loggerLvl:         "",
			syslogLvl:         "",
			expectedLoggerLvl: slog.LevelInfo,
			expectedSyslogLvl: slog.LevelInfo,
		},
		{
			name:              "BothSet",
			loggerLvl:         "warn",
			syslogLvl:         "error",
			expectedLoggerLvl: slog.LevelWarn,
			expectedSyslogLvl: slog.LevelError,
		},
		{
			name:              "SyslogSet",
			loggerLvl:         "",
			syslogLvl:         "debug",
			expectedLoggerLvl: slog.LevelInfo,
			expectedSyslogLvl: slog.LevelDebug,
		},
		{
			name:              "LogSet",
			loggerLvl:         "debug",
			syslogLvl:         "",
			expectedLoggerLvl: slog.LevelDebug,
			expectedSyslogLvl: slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.loggerLvl != "" {
				t.Setenv("IMGPROXY_LOG_LEVEL", tt.loggerLvl)
			}
			if tt.syslogLvl != "" {
				t.Setenv("IMGPROXY_SYSLOG_LEVEL", tt.syslogLvl)
			}

			cfg := NewDefaultConfig()
			LoadConfigFromEnv(&cfg)

			require.Equal(t, tt.expectedLoggerLvl, cfg.Level)
			require.Equal(t, tt.expectedSyslogLvl, cfg.Syslog.Level)
		})
	}
}
