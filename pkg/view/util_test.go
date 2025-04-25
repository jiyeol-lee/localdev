package view

import "testing"

func Test_convertCommandKeyToCharacter(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "lowercase a",
			args: args{
				key: "lowerA",
			},
			want:    "a",
			wantErr: false,
		},
		{
			name: "uppercase a",
			args: args{
				key: "upperA",
			},
			want:    "A",
			wantErr: false,
		},
		{
			name: "invalid",
			args: args{
				key: "invalidKey",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "lowercase c",
			args: args{
				key: "lowerC",
			},
			want:    "c",
			wantErr: false,
		},
		{
			name: "upper c",
			args: args{
				key: "upperC",
			},
			want:    "C",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertCommandKeyToCharacter(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertCommandKeyToCharacter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("convertCommandKeyToCharacter() = %v, want %v", got, tt.want)
			}
		})
	}
}
