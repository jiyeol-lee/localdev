package util

import (
	"reflect"
	"testing"
)

func TestYamlToMap(t *testing.T) {
	type inputData struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}

	tests := []struct {
		name    string
		data    inputData
		want    map[string]any
		wantErr bool
	}{
		{
			name: "simple struct conversion",
			data: inputData{Name: "John", Age: 30},
			want: map[string]any{
				"name": "John",
				"age":  uint64(30),
			},
			wantErr: false,
		},
		{
			name: "empty struct conversion",
			data: inputData{},
			want: map[string]any{
				"name": "",
				"age":  uint64(0),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := YamlToMap[inputData, any](tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("YamlToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("YamlToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
