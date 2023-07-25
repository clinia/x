package errorx

import (
	"encoding/json"
	"fmt"
)

type ErrorType int

const (
	// The Invalid type should not be used, only useful to assert whether or not an error is an MdmError during cast
	ErrorTypeUnspecified ErrorType = iota
	ErrorTypeInternal
	ErrorTypeOutOfRange
	ErrorTypeUnsupported // Bad request
	ErrorTypeNotFound
	ErrorTypeAlreadyExists // Conflict
	// Map to 422
	ErrorTypeInvalidFormat
	// Map to 400
	ErrorTypeFailedPrecondition
)

// TODO: update below
var (
	ErrorType_name = map[int]string{
		0: "UNSPECIFIED",
		1: "INTERNAL",
		2: "OUT_OF_RANGE",
		3: "UNSUPPORTED",
		4: "NOT_FOUND",
		5: "ALREADY_EXISTS",
		6: "INVALID_FORMAT",
		7: "FAILED_PRECONDITION",
	}
	ErrorType_value = map[string]int{
		"UNSPECIFIED":         0,
		"INTERNAL":            1,
		"OUT_OF_RANGE":        2,
		"UNSUPPORTED":         3,
		"NOT_FOUND":           4,
		"ALREADY_EXISTS":      5,
		"INVALID_FORMAT":      6,
		"FAILED_PRECONDITION": 7,
	}
)

func (u ErrorType) String() string {
	return ErrorType_name[int(u)]
}

func (e ErrorType) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func ParseErrorType(s string) (ErrorType, error) {
	value, ok := ErrorType_value[s]
	if !ok {
		return ErrorTypeUnspecified, fmt.Errorf("%q is not a valid error type", s)
	}

	return ErrorType(value), nil
}

func (e *ErrorType) UnmarshalJSON(data []byte) (err error) {
	var errorType string
	if err := json.Unmarshal(data, &errorType); err != nil {
		return err
	}

	if *e, err = ParseErrorType(errorType); err != nil {
		return err
	}

	return nil
}
