package utils

import (
	"testing"
)

func TestIsNumber(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TestIsNumber",
			args: args{
				s: "123",
			},
			want: true,
		},
		{
			name: "TestIsNumber",
			args: args{
				s: "abc",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNumber(tt.args.s); got != tt.want {
				t.Errorf("IsNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasDomain(t *testing.T) {
	testCases := []struct {
		url      string
		expected bool
	}{
		{"https://www.example.com", true},
		{"https://www.facebook.com", true},
		{"http://google.com", true},
		{"https://docs.google.co.uk", true},
		{"https://example.com", true},
		{"https://example.co.uk", true},
		{"https://sub.example.co.uk", true},
		{"https://sub.sub.example.co.uk", true},
		{"https://example..com", false},
		{"ftp://example.com", true},
		{"not a url", false},
	}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			result := HasDomain(tc.url)
			if result != tc.expected {
				t.Errorf("HasDomain() = %v, want %v", result, tc.expected)
			}
		})
	}
}
