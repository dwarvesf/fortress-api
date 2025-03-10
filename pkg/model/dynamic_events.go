package model

import (
	"encoding/json"
	"time"
)

// DynamicEvent represents the event to be stored in parquet format
type DynamicEvent struct {
	Data      json.RawMessage `parquet:"name=data, type=BYTE_ARRAY, convertedtype=UTF8"`
	EventType string          `parquet:"name=event_type, type=BYTE_ARRAY, convertedtype=UTF8"`
	Timestamp time.Time       `parquet:"name=timestamp, type=INT96"`
}
