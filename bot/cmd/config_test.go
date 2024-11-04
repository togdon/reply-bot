package main

import (
	"reflect"
	"testing"
)

func TestReadConfigFromEnv(t *testing.T) {
	tests := []struct {
		name string
		want map[string]string
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readConfigFromEnv(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readConfigFromENV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
