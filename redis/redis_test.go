package redis

import (
	"testing"
	"time"
)

func Test_set(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test_1",
			args: args{
				key:   "deviceLimit",
				value: "xxx",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, err := New([]string{""}, "")
			if err != nil {
				t.Error(err)
				return
			}
			if err := cli.Set(tt.args.key, tt.args.value, 10*time.Second); err != nil {
				t.Errorf("getCellNum() err %v", err)
			}
		})
	}
}
