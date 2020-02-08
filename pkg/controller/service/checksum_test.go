package service

import (
	"testing"
)

type test struct {
	Value string `json:"value"`
}

func TestChecksum(t *testing.T) {
	type args struct {
		spec interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple_string",
			args: args{
				spec: &test{
					Value: "1324",
				},
			},
			want:    "114b305aea2dc44abb98d0d566131733",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Checksum(tt.args.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("Checksum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Checksum() got = %v, want %v", got, tt.want)
			}
		})
	}
}
