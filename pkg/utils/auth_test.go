package utils

import (
	"testing"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func TestGenerateJWTToken(t *testing.T) {
	type args struct {
		info      *model.AuthenticationInfo
		expiresAt int64
		secretKey string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				info: &model.AuthenticationInfo{
					UserID: "2655832e-f009-4b73-a535-64c3a22e558f",
				},
				expiresAt: 4823407476,
				secretKey: "JWTSecretKey",
			},
			want:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ4MjM0MDc0NzYsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiIiwiZW1haWwiOiIifQ.lPYfCDdI7J7lgCEzGLr3xEL80AzcCQ3KtmeIrsEkuB4",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateJWTToken(tt.args.info, tt.args.expiresAt, tt.args.secretKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateJWTToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateJWTToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserIDFromToken(t *testing.T) {
	type args struct {
		tokenString string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				tokenString: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ4MjM0MDc0NzYsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiIiwiZW1haWwiOiIifQ.lPYfCDdI7J7lgCEzGLr3xEL80AzcCQ3KtmeIrsEkuB4",
			},
			want:    "2655832e-f009-4b73-a535-64c3a22e558f",
			wantErr: false,
		},
		{
			name: "invalid token case",
			args: args{
				tokenString: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwLCJpZCI6IjI2NTU4MzJlLWYwMDktNGI3My1hNTM1LTY0YzNhMjJlNTU4ZiIsImF2YXRhciI6IiIsImVtYWlsIjoiIn0.jGnI2bJks8uYV0Siwt-NFj-RpC2ZdHBiAG-iVBZWfLU",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserIDFromToken(tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserIDFromToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetUserIDFromToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
