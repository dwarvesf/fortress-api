package store

import (
	"errors"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/metrics"
)

// FinallyFunc function to finish a transaction
type FinallyFunc = func(error) error

// DBRepo ..
type DBRepo interface {
	DB() *gorm.DB
	NewTransaction() (DBRepo, FinallyFunc)
	SetNewDB(*gorm.DB)
}

// repo is implementation of repository
type repo struct {
	Database *gorm.DB
}

// DB database connection
func (s *repo) DB() *gorm.DB {
	return s.Database
}

func NewRepo(db *gorm.DB) DBRepo {
	return &repo{Database: db}
}

// NewTransaction for database connection
func (s *repo) NewTransaction() (newRepo DBRepo, finallyFn FinallyFunc) {
	newDB := s.Database.Begin()

	finallyFn = func(err error) error {
		if err != nil {
			nErr := newDB.Rollback().Error
			// Track rollback metric
			metrics.DatabaseTransactions.WithLabelValues("rollback").Inc()
			if nErr != nil {
				return errors.New(nErr.Error())
			}
			return err
		}

		cErr := newDB.Commit().Error
		if cErr != nil {
			// Track rollback metric for failed commits
			metrics.DatabaseTransactions.WithLabelValues("rollback").Inc()
			return errors.New(cErr.Error())
		}
		// Track successful commit metric
		metrics.DatabaseTransactions.WithLabelValues("commit").Inc()
		return nil
	}

	return &repo{Database: newDB}, finallyFn
}

func (s *repo) SetNewDB(db *gorm.DB) {
	s.Database = db
}
