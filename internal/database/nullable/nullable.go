package nullable

import "time"

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

// Deref dereferences a pointer, returning the zero value if nil.
func Deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

// StrPtr returns a pointer to the string, or nil if empty.
func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// TimePtr returns the time pointer as-is (pass-through helper).
func TimePtr(t *time.Time) *time.Time {
	return t
}

// TimePtrFromValue returns a pointer to the time, or nil if zero.
func TimePtrFromValue(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// Int64Ptr converts an int to an *int64.
func Int64Ptr(i int) *int64 {
	v := int64(i)
	return &v
}

// DerefInt64 dereferences an *int64 to int, returning 0 if nil.
func DerefInt64(i *int64) int {
	if i == nil {
		return 0
	}
	return int(*i)
}

// BoolToInt64Ptr converts a bool to *int64 (0 or 1).
func BoolToInt64Ptr(b bool) *int64 {
	v := int64(0)
	if b {
		v = 1
	}
	return &v
}

// DerefInt64Bool dereferences an *int64 to bool (non-zero = true).
func DerefInt64Bool(i *int64) bool {
	if i == nil {
		return false
	}
	return *i != 0
}

// ToFloat64 converts various numeric types to float64.
func ToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case int:
		return float64(val)
	default:
		return 0
	}
}

// ToInt converts various numeric types to int.
func ToInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int64:
		return int(val)
	case int:
		return val
	default:
		return 0
	}
}
