package payroll

// func TestGetList(t *testing.T) {
// 	db, _, close := util.GetTestDB(t)
// 	defer close()

// 	var payrolls []domain.Payroll
// 	firstDueDate, err := timeutil.ParseStringToDate("2019-15-08")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 		return
// 	}
// 	secondDueDate, err := timeutil.ParseStringToDate("2019-01-08")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 		return
// 	}
// 	firstUser := domain.User{
// 		Model: domain.Model{
// 			ID: domain.MustGetUUIDFromString("63d163a7-e9f5-4210-a685-151061fe9c29"),
// 		},
// 		BaseSalary: domain.BaseSalary{
// 			Model: domain.Model{
// 				ID: domain.MustGetUUIDFromString("e8ccf0b5-f325-43f2-bc0f-717e9e8e0506"),
// 			},
// 		},
// 	}
// 	// secondUser := domain.User{
// 	// 	Model: domain.Model{
// 	// 		ID: domain.MustGetUUIDFromString("d8a6af04-9e0a-4724-97c4-a78ecd5e9bc4"),
// 	// 	},
// 	// 	BaseSalary: domain.BaseSalary{
// 	// 		Model: domain.Model{
// 	// 			ID: domain.MustGetUUIDFromString("118c36e5-5cd7-4984-a82d-4ad0eb05fee9"),
// 	// 		},
// 	// 	},
// 	// }
// 	payrolls = append(payrolls, domain.Payroll{
// 		ID:                 domain.MustGetUUIDFromString("0c64d4da-ab09-41db-9045-fa42ee35682a"),
// 		UserID:             domain.MustGetUUIDFromString("63d163a7-e9f5-4210-a685-151061fe9c29"),
// 		Total:              20000,
// 		Month:              8,
// 		Year:               2019,
// 		CommissionAmount:   0,
// 		ProjectBonusAmount: 0,
// 		DueDate:            secondDueDate,
// 		User:               firstUser,
// 	}, domain.Payroll{
// 		ID:                 domain.MustGetUUIDFromString("7744a878-3cc1-4290-8aef-13bca71a22a5"),
// 		UserID:             domain.MustGetUUIDFromString("63d163a7-e9f5-4210-a685-151061fe9c29"),
// 		Total:              0,
// 		Month:              8,
// 		Year:               2019,
// 		CommissionAmount:   0,
// 		ProjectBonusAmount: 0,
// 		DueDate:            firstDueDate,
// 		User:               firstUser,
// 	})

// 	testcases := []struct {
// 		name    string
// 		q       Query
// 		wantRes []domain.Payroll
// 		wantErr error
// 	}{
// 		{
// 			name:    "case success",
// 			wantRes: payrolls,
// 		},
// 	}
// 	for _, tc := range testcases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			s := &payrollService{}
// 			res, err := s.GetList(db, tc.q)
// 			if err != nil && err != tc.wantErr {
// 				t.Errorf("[payrollStore].GetList() want error: %v, got error: %v", tc.wantErr, err)
// 				return
// 			}
// 			t.Log(len(res))
// 			for i := 0; i < 2; i++ {
// 				if res[i].Total != tc.wantRes[i].Total {
// 					t.Errorf("[payrollStore].GetList() want total response: %v, got total response: %v", tc.wantRes[i].Total, res[i].Total)
// 					return
// 				}
// 				if res[i].CommissionAmount != tc.wantRes[i].CommissionAmount {
// 					t.Errorf("[payrollStore].GetList() want commission amount response: %v, got commission amount response: %v", tc.wantRes[i].CommissionAmount, res[i].CommissionAmount)
// 					return
// 				}
// 				if res[i].ProjectBonusAmount != tc.wantRes[i].ProjectBonusAmount {
// 					t.Errorf("[payrollStore].GetList() want project bonus response: %v, got project bonus response: %v", tc.wantRes[i].ProjectBonusAmount, res[i].ProjectBonusAmount)
// 					return
// 				}
// 				if reflect.DeepEqual(tc.wantRes[i].DueDate, res[i].DueDate) {
// 					t.Errorf("[payrollStore].GetList() want due_date response: %v, got due_date response: %v", tc.wantRes[i].DueDate, res[i].DueDate)
// 					return
// 				}
// 				if res[i].User.ID != tc.wantRes[i].User.ID {
// 					t.Errorf("[payrollStore].GetList() want user ID response: %v, got user ID response: %v", tc.wantRes[i].User.ID, res[i].User.ID)
// 				}
// 				if res[i].User.BaseSalary.ID != tc.wantRes[i].User.BaseSalary.ID {
// 					t.Errorf("[payrollStore].GetList() want base salary ID response: %v, got base salary ID response: %v", tc.wantRes[i].User.BaseSalary.ID, res[i].User.BaseSalary.ID)
// 				}
// 			}
// 		})
// 	}
// }

// func TestUpdateSpecificFields(t *testing.T) {
// 	db, _, close := util.GetTestDB(t)
// 	defer close()

// 	fields := map[string]interface{}{
// 		"total": 10000000000,
// 	}
// 	testcases := []struct {
// 		name    string
// 		id      string
// 		wantErr error
// 		fields  map[string]interface{}
// 	}{
// 		{
// 			name:   "case success",
// 			id:     "7744a878-3cc1-4290-8aef-13bca71a22a5",
// 			fields: fields,
// 		},
// 		{
// 			name:    "case not found",
// 			id:      "7744a878-3cc1-4290-8aef-13bca71a22a4",
// 			fields:  fields,
// 			wantErr: gorm.ErrRecordNotFound,
// 		},
// 	}
// 	for _, tc := range testcases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			s := &payrollService{}
// 			err := s.UpdateSpecificFields(db, tc.id, tc.fields)
// 			if err != nil && err != tc.wantErr {
// 				t.Errorf("[payrollStore].UpdateSpecificFields() want error: %v, got error: %v", tc.wantErr, err)
// 				return
// 			}
// 			if err != nil && tc.wantErr == nil {
// 				t.Error("[payrollStore].UpdateSpecificFields() want error not nil, got error nil")
// 			}
// 		})
// 	}

// }
