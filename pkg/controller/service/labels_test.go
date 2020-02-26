package service

import (
	"reflect"
	"testing"
)

func Test_mergeLabels(t *testing.T) {
	type args struct {
		labels1 map[string]string
		labels2 map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "empty",
			args: args{
				labels1: map[string]string{},
				labels2: map[string]string{},
			},
			want: map[string]string{},
		},
		{
			name: "single1",
			args: args{
				labels1: map[string]string{
					"test1234": "test1234",
				},
				labels2: map[string]string{},
			},
			want: map[string]string{
				"test1234": "test1234",
			},
		},
		{
			name: "single2",
			args: args{
				labels1: map[string]string{},
				labels2: map[string]string{
					"test1234": "test1234",
				},
			},
			want: map[string]string{
				"test1234": "test1234",
			},
		},
		{
			name: "simple",
			args: args{
				labels1: map[string]string{
					"test1": "test1",
					"test2": "test2",
				},
				labels2: map[string]string{
					"test2": "test2",
					"test3": "test3",
				},
			},
			want: map[string]string{
				"test1": "test1",
				"test2": "test2",
				"test3": "test3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeLabels(tt.args.labels1, tt.args.labels2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
