package names

import (
	"testing"
)

func TestFormatDash(t *testing.T) {
	tests := []struct {
		name string
		give string
		want string
	}{
		{
			name: "no dashes",
			give: "a.b.c.d",
			want: "a-b-c-d",
		},
		{
			name: "with dashes",
			give: "a.b--bb.c.d",
			want: "a-b-bb-c-d",
		},
		{
			name: "simple",
			give: "a",
			want: "a",
		},
		{
			name: "simple2",
			give: "a-b",
			want: "a-b",
		},
		{
			name: "simple3",
			give: "a.b",
			want: "a-b",
		},
		{
			name: "simple4",
			give: "a!b",
			want: "a-b",
		},
		{
			name: "simple5",
			give: "a--b",
			want: "a-b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatDash(tt.give); got != tt.want {
				t.Errorf("FormatDash() = %v, want %v", got, tt.want)
			}
		})
	}
}
