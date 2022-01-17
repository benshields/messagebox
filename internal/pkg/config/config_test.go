package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cases := []struct {
		name      string
		req       string
		wantErr   bool
		wantErrAs interface{}
		wantCfg   *Configuration
	}{
		{
			name:      "Fail on nonexistent config path",
			req:       "does/not/exist",
			wantErr:   true,
			wantErrAs: &ConfigError{},
			wantCfg:   nil,
		},
		{
			name:      "Fail on nonexistent config file",
			req:       "config/doesNotExist.yaml",
			wantErr:   true,
			wantErrAs: &ConfigError{},
			wantCfg:   nil,
		},
		{
			name:    "Success with no config path",
			req:     "",
			wantErr: false,
			wantCfg: &Configuration{
				Logger: LoggerConfiguration{
					Preset: "development",
				},
				Server: ServerConfiguration{
					Host: "0.0.0.0",
					Port: "8080",
				},
				Database: DatabaseConfiguration{
					DatabaseName: "messagebox",
					User:         "messagebox_user",
					Password:     "insecure",
					Host:         "0.0.0.0",
					Port:         "5432",
					MaxOpenConns: 50,
					MaxIdleConns: 2,
				},
			},
		},
		{
			name:    "Success with config path",
			req:     "config/default.yaml",
			wantErr: false,
			wantCfg: &Configuration{
				Logger: LoggerConfiguration{
					Preset: "development",
				},
				Server: ServerConfiguration{
					Host: "0.0.0.0",
					Port: "8080",
				},
				Database: DatabaseConfiguration{
					DatabaseName: "messagebox",
					User:         "messagebox_user",
					Password:     "insecure",
					Host:         "0.0.0.0",
					Port:         "5432",
					MaxOpenConns: 50,
					MaxIdleConns: 2,
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			gotCfg, gotErr := New(tt.req)
			if tt.wantErr {
				assert.ErrorAs(t, gotErr, tt.wantErrAs)
			} else {
				assert.NoError(t, gotErr)
			}
			assert.Equal(t, tt.wantCfg, gotCfg)
		})
	}
}
