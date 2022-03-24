package utils

import "testing"

func TestProtectMobile(t *testing.T) {
	type args struct {
		mobile string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test_protect_mobile",
			args: args{
				mobile: "13581695970",
			},
			want: "135******70",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProtectMobile(tt.args.mobile); got != tt.want {
				t.Errorf("ProtectMobile() = %v, want %v", got, tt.want)
			}
		})
	}
}
