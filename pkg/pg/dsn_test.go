package pg

import (
	"net/url"
	"strings"
	"testing"
)

func TestDSN(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   string
	}{
		{
			name: "With SSLMode",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
				DBName:   "mydb",
				SSLMode:  "require",
			},
			want: "postgres://user:pass@localhost:5432/mydb?sslmode=require",
		},
		{
			name: "Without SSLMode",
			config: Config{
				Host:     "127.0.0.1",
				Port:     5433,
				User:     "admin",
				Password: "secret",
				DBName:   "testdb",
				SSLMode:  "",
			},
			want: "postgres://admin:secret@127.0.0.1:5433/testdb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.DSN()
			// Parse resulting URL and expected URL for comparison
			gotURL, err := url.Parse(got)
			if err != nil {
				t.Fatalf("failed to parse DSN: %v", err)
			}
			wantURL, err := url.Parse(tt.want)
			if err != nil {
				t.Fatalf("failed to parse expected DSN: %v", err)
			}

			// Compare scheme, host and path
			if gotURL.Scheme != wantURL.Scheme {
				t.Errorf("scheme mismatch: got %s, want %s", gotURL.Scheme, wantURL.Scheme)
			}
			if gotURL.Host != wantURL.Host {
				t.Errorf("host mismatch: got %s, want %s", gotURL.Host, wantURL.Host)
			}
			if !strings.HasPrefix(gotURL.Path, wantURL.Path) {
				t.Errorf("path mismatch: got %s, want %s", gotURL.Path, wantURL.Path)
			}

			// Compare query parameters
			gotQuery := gotURL.Query()
			wantQuery := wantURL.Query()
			if len(gotQuery) != len(wantQuery) {
				t.Errorf("query parameters length mismatch: got %v, want %v", gotQuery, wantQuery)
			}
			for key, wantValues := range wantQuery {
				gotValues, exists := gotQuery[key]
				if !exists {
					t.Errorf("missing query key: %s", key)
				}
				if len(gotValues) != len(wantValues) {
					t.Errorf("query parameter %s length mismatch: got %v, want %v", key, gotValues, wantValues)
				}
				for i, v := range wantValues {
					if gotValues[i] != v {
						t.Errorf("query parameter %s mismatch: got %s, want %s", key, gotValues[i], v)
					}
				}
			}
		})
	}
}
