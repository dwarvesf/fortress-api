package employeecommission

// func TestCreate(t *testing.T) {
// 	t.Parallel()

// 	db, _, close := util.GetTestDB(t)
// 	defer close()

// 	var cs []domain.UserCommission
// 	cs = append(cs, domain.UserCommission{
// 		UserID:    domain.NewUUID(),
// 		InvoiceID: domain.NewUUID(),
// 		Amount:    1000,
// 	})
// 	testcases := []struct {
// 		name    string
// 		cs      []domain.UserCommission
// 		wantErr error
// 	}{
// 		{
// 			name: "case success",
// 			cs: cs,
// 			wantErr: nil,
// 		},
// 	}
// 	for _, tc := range testcases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			s := &pgStore{}
// 			err := s.Create(db, tc.cs)
// 			if err == nil && tc.wantErr != nil {
// 				t.Errorf("[CommissionStore].Create() want error not nil, got error: %v", err)
// 				return
// 			}
// 			if err != nil && err != tc.wantErr {
// 				t.Errorf("[CommissionStore].Create() want error: %v, got error: %v", tc.wantErr, err)
// 			}
// 		})
// 	}
// }

// func TestGet(t *testing.T) {
// 	db, _, close := util.GetTestDB(t)
// 	defer close()

// 	fromDate := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
// 	toDate := time.Date(2019, 2, 15, 0, 0, 0, 0, time.UTC)

// 	var res []domain.UserCommission
// 	res = append(res, domain.UserCommission{
// 		Model: domain.Model{
// 			ID: domain.MustGetUUIDFromString("a8d67ba6-50ae-45f3-9a67-56469dadc2c3"),
// 		},
// 		Amount:    0,
// 		UserID:    domain.MustGetUUIDFromString("d8a6af04-9e0a-4724-97c4-a78ecd5e9bc4"),
// 		InvoiceID: domain.MustGetUUIDFromString("be0d0584-867e-44e6-b7b7-366e4704e7cb"),
// 		Project:   "Example",
// 	})
// 	testcases := []struct {
// 		name    string
// 		q       Query
// 		wantRes []domain.UserCommission
// 		wantErr error
// 	}{
// 		{
// 			name: "case success",
// 			q: Query{
// 				UserID:   "",
// 				FromDate: &fromDate,
// 				ToDate:   &toDate,
// 			},
// 			wantRes: res,
// 		},
// 	}
// 	for _, tc := range testcases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			s := &pgStore{}
// 			res, err := s.Get(db, tc.q)
// 			if err != nil && err != tc.wantErr {
// 				t.Errorf("[commissionStore].Get() want error: %v, got error: %v", tc.wantErr, err)
// 				return
// 			}
// 			if err == nil {
// 				for i := range res {
// 					if res[i].Amount != tc.wantRes[i].Amount {
// 						t.Errorf("[commissionStore].Get() want amount: %v, got amount: %v", tc.wantRes[i].Amount, res[i].Amount)
// 						return
// 					}
// 					if res[i].UserID != tc.wantRes[i].UserID {
// 						t.Errorf("[commissionStore].Get() want user ID: %v, got user ID: %v", tc.wantRes[i].UserID, res[i].UserID)
// 						return
// 					}
// 					if res[i].InvoiceID != tc.wantRes[i].InvoiceID {
// 						t.Errorf("[commissionStore].Get() want invoice ID: %v, got invoice ID: %v", tc.wantRes[i].InvoiceID, res[i].InvoiceID)
// 						return
// 					}
// 					if res[i].Project != tc.wantRes[i].Project {
// 						t.Errorf("[commissionStore].Get() want project: %v, got project: %v", tc.wantRes[i].Project, res[i].Project)
// 						return
// 					}
// 					if res[i].ID != tc.wantRes[i].ID {
// 						t.Errorf("[commissionStore].Get() want ID: %v, got ID: %v", tc.wantRes[i].ID, res[i].ID)
// 					}
// 				}
// 			}
// 		})
// 	}
// }
