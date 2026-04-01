package config

import (
	"testing"
)

func TestConvertURIToDSN(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		want    string
		wantErr bool
	}{
		{
			name: "Standard Railway URI",
			uri:  "mysql://root:password@localhost:3306/e-plan-ai",
			want: "root:password@tcp(localhost:3306)/e-plan-ai",
		},
		{
			name: "URI with Query Parameters",
			uri:  "mysql://user:pass@remote-host:3307/db_name?charset=utf8mb4&parseTime=True",
			want: "user:pass@tcp(remote-host:3307)/db_name?charset=utf8mb4&parseTime=True",
		},
		{
			name: "URI without Port",
			uri:  "mysql://root:secret@localhost/test_db",
			want: "root:secret@tcp(localhost:3306)/test_db",
		},
		{
			name: "Non-MySQL URI",
			uri:  "postgres://user:pass@host/db",
			want: "postgres://user:pass@host/db", // Should return as is
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertURIToDSN(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertURIToDSN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("convertURIToDSN() got = %v, want %v", got, tt.want)
			}
		})
	}
}
