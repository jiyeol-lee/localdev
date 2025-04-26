package view

import "testing"

func Test_keyToFocusAction(t *testing.T) {
	type args struct {
		key             rune
		textViewsLength int
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 bool
	}{
		{
			name: "If key is 1 and textViewsLength is 5",
			args: args{
				key:             49,
				textViewsLength: 5,
			},
			want:  0,
			want1: true,
		},
		{
			name: "If key is 1 and textViewsLength is 0",
			args: args{
				key:             49,
				textViewsLength: 0,
			},
			want:  0,
			want1: false,
		},
		{
			name: "If key is 2 and textViewsLength is 5",
			args: args{
				key:             50,
				textViewsLength: 5,
			},
			want:  1,
			want1: true,
		},
		{
			name: "If key is 2 and textViewsLength is 1",
			args: args{
				key:             50,
				textViewsLength: 1,
			},
			want:  1,
			want1: false,
		},
		{
			name: "If key is not 0-9 and textViewsLength is 10",
			args: args{
				key:             58,
				textViewsLength: 10,
			},
			want:  -1,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := keyToFocusAction(tt.args.key, tt.args.textViewsLength)
			if got != tt.want {
				t.Errorf("keyToFocusAction() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("keyToFocusAction() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
