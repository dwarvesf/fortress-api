package store

import (
	"gorm.io/gorm"
)

// testRepo is implementation of repository
type testRepo struct {
	Database *gorm.DB
}

// DB database connection
func (s *testRepo) DB() *gorm.DB {
	return s.Database
}

func NewTestRepo(db *gorm.DB) DBRepo {
	return &testRepo{Database: db}
}

// NewTransaction for database connection
func (s *testRepo) NewTransaction() (newRepo DBRepo, finallyFn FinallyFunc) {
	finallyFn = func(err error) error {
		return err
	}

	return &testRepo{Database: s.DB()}, finallyFn
}

func (s *testRepo) SetNewDB(db *gorm.DB) {
	s.Database = db
}
