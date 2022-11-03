package model

import "testing"

func Test_StandardizeSortQuery(t *testing.T) {
	type args struct {
		sortQ string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				sortQ: "",
			},
			want: "",
		},
		{
			name: "no sort operator",
			args: args{
				sortQ: "chain_id",
			},
			want: "chain_id ASC",
		},
		{
			name: "multiple empty sort",
			args: args{
				sortQ: ",,,",
			},
			want: "",
		},
		{
			name: "multiple sort",
			args: args{
				sortQ: "-token_id,name",
			},
			want: "token_id DESC,name ASC",
		},
		{
			name: "invalid sort_operator/field",
			args: args{
				sortQ: "token_id?",
			},
			want: "token_id ASC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := standardizeSortQuery(tt.args.sortQ); got != tt.want {
				t.Errorf("StandardizeUri() = %v, want %v", got, tt.want)
			}
		})
	}
}
