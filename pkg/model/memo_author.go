package model

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MemoAuthor is the join table for memo log and community member
type MemoAuthor struct {
	MemoLogID         UUID `gorm:"primaryKey"`
	CommunityMemberID UUID `gorm:"primaryKey"`
	CreatedAt         time.Time
}

func (b *MemoAuthor) BeforeCreate(tx *gorm.DB) (err error) {
	cols := []clause.Column{}
	for _, field := range tx.Statement.Schema.PrimaryFields {
		cols = append(cols, clause.Column{Name: field.DBName})
	}
	tx.Statement.AddClause(clause.OnConflict{
		Columns:   cols,
		DoNothing: true,
	})

	return nil
}
