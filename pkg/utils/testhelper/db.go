package testhelper

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

var (
	db              *gorm.DB
	fixtures        *testfixtures.Loader
	singletonTestDB sync.Once
)

func LoadTestDB() *gorm.DB {
	var err error
	var conn *sql.DB

	singletonTestDB.Do(func() {
		// initiate logger
		l := logger.NewLogrusLogger()

		conn, err = sql.Open("postgres", "host=localhost port=35432 user=postgres password=postgres dbname=fortress_local_test sslmode=disable")
		if err != nil {
			l.Fatalf(err, "failed to open database connection")
			return
		}

		path, err := os.Getwd()
		if err != nil {
			l.Error(err, "unable to get dir")
		}
		fmt.Println(path) // for example /home/user

		// load fixture and restore db
		fixtures, err = testfixtures.New(
			testfixtures.Database(conn),
			testfixtures.Dialect("postgres"),
			testfixtures.Directory("../../../migrations/test_seed"),
		)
		if err != nil {
			l.Fatalf(err, "failed to load fixture")
			return
		}

		if err = fixtures.Load(); err != nil {
			l.Fatalf(err, "failed to load fixture")
			return
		}

		db, err = gorm.Open(postgres.New(
			postgres.Config{Conn: conn}),
			&gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					SingularTable: false,
				},
			})
		if err != nil {
			l.Fatalf(err, "gorm: failed to open database connection")
		}
	})

	return db
}

func TestWithTxDB(t *testing.T, callback func(tx store.DBRepo)) {
	var appDB store.DBRepo
	var err error
	var conn *sql.DB

	conn, err = sql.Open("postgres", "host=localhost port=35432 user=postgres password=postgres dbname=fortress_local_test sslmode=disable")
	require.NoError(t, err)
	db, err := gorm.Open(postgres.New(
		postgres.Config{Conn: conn}),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: false,
			},
		})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	appDB = store.NewTestRepo(db)
	newTx := appDB.DB().Begin()
	defer newTx.Rollback()
	appDB.SetNewDB(newTx)
	callback(appDB)
}

func LoadTestSQLFile(t *testing.T, txRepo store.DBRepo, fileName string) {
	file, err := os.ReadFile(fileName)
	require.NoError(t, err)
	for _, q := range strings.Split(string(file), ";") {
		q := strings.TrimSpace(q)
		if q == "" {
			continue
		}
		err = txRepo.DB().Exec(q).Error
		require.NoError(t, err)
	}
}
