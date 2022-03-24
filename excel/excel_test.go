package excel

import "testing"

func Test_getCellNum(t *testing.T) {
	type args struct {
		row    int
		column int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test_1",
			args: args{
				row:    3,
				column: 6,
			},
			want: "D6",
		},
		{
			name: "test_2",
			args: args{
				row:    27,
				column: 6,
			},
			want: "AB6",
		},
		{
			name: "test_3",
			args: args{
				row:    28,
				column: 6,
			},
			want: "AC6",
		},
		{
			name: "test_4",
			args: args{
				row:    50,
				column: 6,
			},
			want: "AY6",
		},
		{
			name: "test_5",
			args: args{
				row:    52,
				column: 2,
			},
			want: "BA2",
		},
		{
			name: "test_6",
			args: args{
				row:    25,
				column: 2,
			},
			want: "Z2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCellNum(tt.args.row, tt.args.column); got != tt.want {
				t.Errorf("getCellNum() = %v, want %v", got, tt.want)
			}
		})
	}
}
