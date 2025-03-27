package app

import (
	"reflect"
	"testing"
)

func Test_getGridDimensions(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		name     string
		args     args
		wantRows int
		wantCols int
	}{
		{
			name:     "0 length",
			args:     args{length: 0},
			wantRows: 0,
			wantCols: 0,
		},
		{
			name:     "1 length",
			args:     args{length: 1},
			wantRows: 1,
			wantCols: 1,
		},
		{
			name:     "2 length",
			args:     args{length: 2},
			wantRows: 1,
			wantCols: 2,
		},
		{
			name:     "3 length",
			args:     args{length: 3},
			wantRows: 2,
			wantCols: 2,
		},
		{
			name:     "4 length",
			args:     args{length: 4},
			wantRows: 2,
			wantCols: 2,
		},
		{
			name:     "5 length",
			args:     args{length: 5},
			wantRows: 2,
			wantCols: 3,
		},
		{
			name:     "6 length",
			args:     args{length: 6},
			wantRows: 2,
			wantCols: 3,
		},
		{
			name:     "7 length",
			args:     args{length: 7},
			wantRows: 2,
			wantCols: 4,
		},
		{
			name:     "8 length",
			args:     args{length: 8},
			wantRows: 2,
			wantCols: 4,
		},
		{
			name:     "9 length",
			args:     args{length: 9},
			wantRows: 2,
			wantCols: 5,
		},
		{
			name:     "10 length",
			args:     args{length: 10},
			wantRows: 2,
			wantCols: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRows, gotCols := getGridDimensions(tt.args.length)
			if gotRows != tt.wantRows {
				t.Errorf("getGridDimensions() gotRows = %v, want %v", gotRows, tt.wantRows)
			}
			if gotCols != tt.wantCols {
				t.Errorf("getGridDimensions() gotCols = %v, want %v", gotCols, tt.wantCols)
			}
		})
	}
}

func Test_makeFlexibleSlice(t *testing.T) {
	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "size 0",
			args: args{size: 0},
			want: []int{},
		},
		{
			name: "size 1",
			args: args{size: 1},
			want: []int{0},
		},
		{
			name: "size 2",
			args: args{size: 2},
			want: []int{0, 0},
		},
		{
			name: "size 10",
			args: args{size: 10},
			want: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeFlexibleSlice(tt.args.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeFlexibleSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
