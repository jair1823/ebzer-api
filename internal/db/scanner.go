package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

// NullTime is a wrapper around sql.NullTime that handles SQLite TEXT timestamps
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Scan implements the sql.Scanner interface
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Valid = false
		return nil
	}

	var err error
	switch v := value.(type) {
	case time.Time:
		nt.Time = v
		nt.Valid = true
	case string:
		// Try multiple datetime formats that SQLite might use
		formats := []string{
			"2006-01-02 15:04:05.999999999-07:00", // SQLite datetime with timezone
			"2006-01-02 15:04:05-07:00",           // SQLite datetime with timezone (no microseconds)
			"2006-01-02 15:04:05.999999999+07:00", // SQLite datetime with + timezone
			"2006-01-02 15:04:05+07:00",           // SQLite datetime with + timezone (no microseconds)
			"2006-01-02T15:04:05.999999999Z07:00", // ISO 8601 with timezone
			time.RFC3339Nano,                      // ISO 8601 nano
			time.RFC3339,                          // ISO 8601
			"2006-01-02 15:04:05.999999999",       // SQLite datetime with microseconds
			"2006-01-02 15:04:05",                 // SQLite datetime
			"2006-01-02",                          // Date only
		}

		for _, format := range formats {
			nt.Time, err = time.Parse(format, v)
			if err == nil {
				nt.Valid = true
				return nil
			}
		}
		return err
	case []byte:
		// Treat []byte same as string
		return nt.Scan(string(v))
	default:
		// Fallback to sql.NullTime
		var sqlNullTime sql.NullTime
		if err := sqlNullTime.Scan(value); err != nil {
			return err
		}
		nt.Time = sqlNullTime.Time
		nt.Valid = sqlNullTime.Valid
	}

	return nil
}

// Value implements the driver.Valuer interface
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// MarshalJSON implements json.Marshaler interface
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (nt *NullTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nt.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &nt.Time); err != nil {
		return err
	}
	nt.Valid = true
	return nil
}

// Time returns a Scanner for non-nullable time.Time fields
type Time struct {
	time.Time
}

// Scan implements the sql.Scanner interface for Time
func (t *Time) Scan(value interface{}) error {
	nt := &NullTime{}
	if err := nt.Scan(value); err != nil {
		return err
	}
	if !nt.Valid {
		// For non-nullable fields, we use zero time instead of error
		t.Time = time.Time{}
		return nil
	}
	t.Time = nt.Time
	return nil
}

// Value implements the driver.Valuer interface
func (t Time) Value() (driver.Value, error) {
	return t.Time, nil
}
