package duration

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration :
// A wrapper around the standard library duration to
// provide custom `JSON` marshalling so that it can
// support other things than nanoseconds.
// This post was useful in order to come up with this
// set of functions:
// https://stackoverflow.com/questions/48050945/how-to-unmarshal-json-into-durations
// This element extends the behavior provided by the
// `time.Duration` object.
type Duration struct {
	time.Duration
}

// ErrInvalidInput :
// Indicates that the value provided as input cannot
// be unmarshalled into a valid duration.
var ErrInvalidInput = fmt.Errorf("could not umarshal value to duration")

// NewDuration :
// Creates a new duration from a base time.Duration.
//
// The `t` defines the wrapped duration.
//
// Returns the created duration.
func NewDuration(t time.Duration) Duration {
	return Duration{
		t,
	}
}

// MarshalJSON :
// Imlepementation of the marshaller interface to be
// able to use this object out-of-the-box with the
// `encoding/json` package provided by the standard
// library.
//
// Returns the marshalled bytes corresponding to this
// object along with any errors.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON :
// Second facet of the Marshaller interface which
// allows to extract the duration from raw bytes.
//
// The `b` defines the bytes to unmarshal.
//
// Returns any error.
func (d *Duration) UnmarshalJSON(b []byte) error {
	// Unmarshal the content using the base encoder.
	// We will then detect which actual datatype is
	// represented by the input bytes and convert it
	// accordingly to a duration.
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	// Convert the value into a meaningful duration.
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return ErrInvalidInput
	}
}
