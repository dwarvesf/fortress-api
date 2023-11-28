package view

import (
	"database/sql/driver"
	"errors"

	"github.com/dwarvesf/fortress-api/pkg/model"
	uuid "github.com/satori/go.uuid"
)

// UUID implement for go-pg convert uuid
type UUID [16]byte // @name UUID

func ToModelUUIDs(ids []UUID) []model.UUID {
	if len(ids) == 0 {
		return nil
	}
	result := make([]model.UUID, 0, len(ids))
	for _, id := range ids {
		result = append(result, model.UUID(id))
	}
	return result
}

// IsZero check uuid is zero
func (u *UUID) IsZero() bool {
	if u == nil {
		return true
	}
	for _, c := range u {
		if c != 0 {
			return false
		}
	}
	return true
}

func (u UUID) String() string {
	if u.IsZero() {
		return ""
	}
	return uuid.UUID(u).String()
}

// MarshalJSON implement for json encoding
func (u UUID) MarshalJSON() ([]byte, error) {
	if len(u) == 0 {
		return []byte(`""`), nil
	}
	return []byte(`"` + u.String() + `"`), nil
}

// UnmarshalJSON implement for json decoding
func (u *UUID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == `""` {
		return nil
	}

	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errors.New("invalid UUID format")
	}
	data = data[1 : len(data)-1]
	uu, err := uuid.FromString(string(data))
	if err != nil {
		return errors.New("invalid UUID format")
	}
	*u = UUID(uu)
	return nil
}

// Value .
func (u UUID) Value() (driver.Value, error) {
	if u.IsZero() {
		return nil, nil
	}
	return uuid.UUID(u).String(), nil
}

func UUIDFromString(s string) (UUID, error) {
	id, err := uuid.FromString(s)
	return UUID(id), err
}

func MustGetUUIDFromString(s string) UUID {
	id, err := uuid.FromString(s)
	if err != nil {
		panic(err)
	}
	return UUID(id)
}
