package script

import (
	"fmt"
	"math"
)

// ValidateJSONValue verifies value contains only JSON-like data that can be
// safely emitted as an object record: nil, bool, string, finite numbers,
// []any, and map[string]any.
func ValidateJSONValue(value any) error {
	return validateJSONValue("$", value)
}

// ValidateRecord verifies one script-produced record is JSON-like.
func ValidateRecord(record map[string]any) error {
	if record == nil {
		return fmt.Errorf("record must not be nil")
	}

	return ValidateJSONValue(record)
}

// ValidateRecords verifies all script-produced records are JSON-like.
func ValidateRecords(records []map[string]any) error {
	for i, record := range records {
		if err := ValidateRecord(record); err != nil {
			return fmt.Errorf("records[%d]: %w", i, err)
		}
	}

	return nil
}

// validateJSONValue recursively validates one value and reports JSON path on error.
func validateJSONValue(path string, value any) error {
	if err, ok := validateScalar(path, value); ok {
		return err
	}

	switch value := value.(type) {
	case []any:
		return validateJSONArray(path, value)
	case map[string]any:
		return validateJSONObject(path, value)
	default:
		return fmt.Errorf("%s: unsupported value type %T", path, value)
	}
}

// validateScalar handles nil, booleans, strings, integers, and finite floats.
func validateScalar(path string, value any) (error, bool) {
	switch value := value.(type) {
	case nil, bool, string,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return nil, true
	case float32:
		return validateFloat(path, float64(value)), true
	case float64:
		return validateFloat(path, value), true
	default:
		return nil, false
	}
}

// validateJSONArray checks every list item recursively.
func validateJSONArray(path string, value []any) error {
	for i, item := range value {
		if err := validateJSONValue(fmt.Sprintf("%s[%d]", path, i), item); err != nil {
			return err
		}
	}

	return nil
}

// validateJSONObject checks every object property recursively.
func validateJSONObject(path string, value map[string]any) error {
	for key, item := range value {
		if err := validateJSONValue(fmt.Sprintf("%s.%s", path, key), item); err != nil {
			return err
		}
	}

	return nil
}

// validateFloat rejects NaN and infinity because JSON cannot encode them.
func validateFloat(path string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("%s: non-finite number", path)
	}

	return nil
}
