package view

import (
	"reflect"
	"testing"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func TestToAuthData(t *testing.T) {
	id := model.NewUUID()
	type args struct {
		accessToken string
		employee    *model.Employee
	}
	tests := []struct {
		name string
		args args
		want *AuthData
	}{
		{
			name: "happy case",
			args: args{
				accessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ4MjM0MDc0NzYsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiIiwiZW1haWwiOiIifQ.RJhl4O1PDMzIFO3pR13vyO07Z4gd90ewq5PKOao1MtY",
				employee: &model.Employee{
					BaseModel: model.BaseModel{
						ID: id,
					},
				},
			},
			want: &AuthData{
				Employee: EmployeeData{
					BaseModel: model.BaseModel{
						ID: id,
					},
					Projects:  []EmployeeProjectData{},
					Positions: []model.Position{},
				},
				AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ4MjM0MDc0NzYsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiIiwiZW1haWwiOiIifQ.RJhl4O1PDMzIFO3pR13vyO07Z4gd90ewq5PKOao1MtY",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToAuthData(tt.args.accessToken, tt.args.employee); !reflect.DeepEqual(got.Employee.ID, tt.want.Employee.ID) {
				t.Errorf("ToAuthData() = %v, want %v", *got, *tt.want)
			}
		})
	}
}
