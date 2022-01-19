package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/benshields/messagebox/internal/pkg/config"
)

func TestSetup(t *testing.T) {
	devLog, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("zap.NewDevelopment() failed with %s", err)
	}
	prodLog, err := zap.NewProduction()
	if err != nil {
		t.Fatalf("zap.NewProduction() failed with %s", err)
	}

	cases := []struct {
		name      string
		req       config.LoggerConfiguration
		wantErr   bool
		wantErrAs interface{}
		wantLog   *zap.Logger
	}{
		{
			name: "Fail on blank preset",
			req: config.LoggerConfiguration{
				Preset: "",
			},
			wantErr:   true,
			wantErrAs: &PresetError{},
			wantLog:   nil,
		},
		{
			name: "Fail on unknown preset",
			req: config.LoggerConfiguration{
				Preset: "unknown",
			},
			wantErr:   true,
			wantErrAs: &PresetError{},
			wantLog:   nil,
		},
		{
			name: "Success development",
			req: config.LoggerConfiguration{
				Preset: development,
			},
			wantErr: false,
			wantLog: devLog,
		},
		{
			name: "Success production",
			req: config.LoggerConfiguration{
				Preset: production,
			},
			wantErr: false,
			wantLog: prodLog,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			gotLog, gotErr := Setup(tt.req)
			if tt.wantErr {
				assert.ErrorAs(t, gotErr, tt.wantErrAs)
			} else {
				assert.NoError(t, gotErr)
			}
			assert.IsType(t, tt.wantLog, gotLog)
		})
	}
}
