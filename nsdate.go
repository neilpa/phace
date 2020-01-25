package phace

import (
	"fmt"
	"time"
)

// NSDate is the serialization type for time values in the photos DB
// https://developer.apple.com/documentation/foundation/nsdate?language=objc
// > NSDate objects encapsulate a single point in time, independent of any
// > particular calendrical system or time zone. Date objects are immutable,
// > representing an invariant time interval relative to an absolute reference
// > date (00:00:00 UTC on 1 January 2001).
type NSDate float64

var (
	epoch = time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)
)

// Scan implements sql.Scanner to load NSDate values. Handles float64, int64
// and time.Time values. The last interprets the `Time.Unix` seconds relative
// to the NSDate epoch.
func (d *NSDate) Scan(src interface{}) error {
	// Can scan as both a time.Time and float64 value
	switch v := src.(type) {
	case time.Time:
		// TODO Loses sub-second precision nanoseconds are also provided
		*d = NSDate(v.Unix())
	case float64:
		*d = NSDate(v)
	case int64:
		*d = NSDate(v)
	default:
		return fmt.Errorf("NSDate: unsupported type: %T", src)
	}
	return nil
}

// Time converts the NSDate to a go time.Time value.
func (d NSDate) Time() time.Time {
	return epoch.Add(time.Second * time.Duration(d))
}

// String returns the UTC RFC3339 representation of the date
func (d NSDate) String() string {
	return d.Time().Format(time.RFC3339)
}
