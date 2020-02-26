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
		{
			name: "null",
			args: args{
				spec: nil,
			},
			want:    "37a6259cc0c1dae299a7866489dff0bd",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checksum(tt.args.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("checksum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checksum() got = %v, want %v", got, tt.want)
			}
		})
	}
}
