package store

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/dwarvesf/fortress-api/pkg/monitoring"
)

// Simple test model for benchmarking
type BenchmarkEmployee struct {
	ID       uint   `gorm:"primarykey"`
	FullName string `gorm:"size:255"`
	Username string `gorm:"size:100;uniqueIndex"`
}

func setupBenchmarkDB(b *testing.B) (*gorm.DB, func()) {
	// Use in-memory SQLite for consistent benchmarking
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
		},
		Logger: nil, // Disable logging for cleaner benchmarks
	})
	require.NoError(b, err)

	// Auto-migrate the test table
	err = db.AutoMigrate(&BenchmarkEmployee{})
	require.NoError(b, err)

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func BenchmarkDatabaseOperationsWithoutMonitoring(b *testing.B) {
	db, cleanup := setupBenchmarkDB(b)
	defer cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create with unique username
		err := db.Create(&BenchmarkEmployee{
			FullName: "Benchmark User",
			Username: fmt.Sprintf("benchuser%d", i),
		}).Error
		if err != nil {
			b.Fatal(err)
		}

		// Query
		var employee BenchmarkEmployee
		err = db.First(&employee, 1).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			b.Fatal(err)
		}

		// Update
		if employee.ID > 0 {
			err = db.Model(&employee).Update("full_name", "Updated Name").Error
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkDatabaseOperationsWithMonitoring(b *testing.B) {
	db, cleanup := setupBenchmarkDB(b)
	defer cleanup()

	// Setup database monitoring
	config := &monitoring.DatabaseMonitoringConfig{
		Enabled:            true,
		CustomMetrics:      true,
		SlowQueryThreshold: 1 * time.Second,
		BusinessMetrics:    true,
	}

	// Register monitoring callbacks
	registerCustomDatabaseCallbacks(db, config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create with unique username
		err := db.Create(&BenchmarkEmployee{
			FullName: "Benchmark User",
			Username: fmt.Sprintf("benchuser_mon_%d", i),
		}).Error
		if err != nil {
			b.Fatal(err)
		}

		// Query
		var employee BenchmarkEmployee
		err = db.First(&employee, 1).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			b.Fatal(err)
		}

		// Update
		if employee.ID > 0 {
			err = db.Model(&employee).Update("full_name", "Updated Name").Error
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkDatabaseBatchOperationsWithoutMonitoring(b *testing.B) {
	db, cleanup := setupBenchmarkDB(b)
	defer cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Batch insert
		employees := make([]BenchmarkEmployee, 10)
		for j := 0; j < 10; j++ {
			employees[j] = BenchmarkEmployee{
				FullName: "Batch Employee",
				Username: fmt.Sprintf("batchuser_%d_%d", i, j),
			}
		}

		err := db.CreateInBatches(employees, 5).Error
		if err != nil {
			b.Fatal(err)
		}

		// Batch query
		var results []BenchmarkEmployee
		err = db.Limit(10).Find(&results).Error
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDatabaseBatchOperationsWithMonitoring(b *testing.B) {
	db, cleanup := setupBenchmarkDB(b)
	defer cleanup()

	// Setup database monitoring
	config := &monitoring.DatabaseMonitoringConfig{
		Enabled:            true,
		CustomMetrics:      true,
		SlowQueryThreshold: 1 * time.Second,
		BusinessMetrics:    true,
	}

	// Register monitoring callbacks
	registerCustomDatabaseCallbacks(db, config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Batch insert
		employees := make([]BenchmarkEmployee, 10)
		for j := 0; j < 10; j++ {
			employees[j] = BenchmarkEmployee{
				FullName: "Batch Employee",
				Username: fmt.Sprintf("batchuser_%d_%d", i, j),
			}
		}

		err := db.CreateInBatches(employees, 5).Error
		if err != nil {
			b.Fatal(err)
		}

		// Batch query
		var results []BenchmarkEmployee
		err = db.Limit(10).Find(&results).Error
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDatabaseMonitoringOverhead(b *testing.B) {
	// This test measures just the monitoring callback overhead
	db, cleanup := setupBenchmarkDB(b)
	defer cleanup()

	config := &monitoring.DatabaseMonitoringConfig{
		Enabled:            true,
		CustomMetrics:      true,
		SlowQueryThreshold: 1 * time.Second,
		BusinessMetrics:    true,
	}

	registerCustomDatabaseCallbacks(db, config)

	// Create a simple record for testing
	employee := BenchmarkEmployee{
		FullName: "Test Employee",
		Username: "testuser",
	}
	db.Create(&employee)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simple query that will trigger monitoring callbacks
		var result BenchmarkEmployee
		db.First(&result, employee.ID)
	}
}

func BenchmarkBusinessDomainInference(b *testing.B) {
	// Benchmark the business domain inference function
	db := &gorm.DB{
		Statement: &gorm.Statement{
			Table: "employees",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		domain := inferBusinessDomain(db)
		_ = domain // Use the result to prevent optimization
	}
}

func BenchmarkTableNameExtraction(b *testing.B) {
	// Benchmark the table name extraction function
	db := &gorm.DB{
		Statement: &gorm.Statement{
			Table: "project_members",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tableName := getTableName(db)
		_ = tableName // Use the result to prevent optimization
	}
}

// TestDatabaseMonitoringPerformanceOverhead measures the performance impact
func TestDatabaseMonitoringPerformanceOverhead(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Run benchmarks programmatically to compare overhead
	withoutMonitoring := testing.Benchmark(BenchmarkDatabaseOperationsWithoutMonitoring)
	withMonitoring := testing.Benchmark(BenchmarkDatabaseOperationsWithMonitoring)

	// Calculate overhead percentage
	overheadNs := withMonitoring.NsPerOp() - withoutMonitoring.NsPerOp()
	overheadPercent := float64(overheadNs) / float64(withoutMonitoring.NsPerOp()) * 100

	t.Logf("Without monitoring: %d ns/op", withoutMonitoring.NsPerOp())
	t.Logf("With monitoring: %d ns/op", withMonitoring.NsPerOp())
	t.Logf("Overhead: %d ns/op (%.2f%%)", overheadNs, overheadPercent)

	// Verify overhead is acceptable (target: < 15% for database operations)
	// Database monitoring has more overhead than HTTP monitoring due to callback complexity
	if overheadPercent > 15.0 {
		t.Errorf("Database monitoring overhead %.2f%% exceeds 15%% target", overheadPercent)
	} else {
		t.Logf("Database monitoring overhead %.2f%% is within acceptable limits", overheadPercent)
	}
}