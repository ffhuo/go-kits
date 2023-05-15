package influx

import (
	"testing"
)

func TestFilterFluxTag(t *testing.T) {
	type args struct {
		str  string
		args []string
	}
	tests := []struct {
		name string
		args args
		want Args
	}{
		{
			name: "test",
			args: args{
				str:  "r.tag1==? or r.tag2==? or r.tag3==?",
				args: []string{"aaa", "bbb", "ccc"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterFluxTag(tt.args.str, tt.args.args...)
			var str string
			got(&str)
			t.Errorf("%v", str)
		})
	}
}
