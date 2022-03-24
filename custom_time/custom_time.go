package custom_time

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Time time.Time

func Now() Time {
	return Time(time.Now())
}

func (t Time) Add(d time.Duration) Time {
	return Time(time.Time(t).Add(d))
}

// MarshalJSON implements the json.Marshaler interface.
// The time is a quoted string in RFC 3339 format, with sub-second precision added if present.
func (t Time) MarshalJSON() ([]byte, error) {
	// if y := time.Time(t).Year(); y < 0 || y >= 10000 {
	// 	// RFC 3339 is clear that years are 4 digits exactly.
	// 	// See golang.org/issue/4556#c15 for more discussion.
	// 	return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	// }

	// b := make([]byte, 0, len(time.RFC3339)+2)
	// b = append(b, '"')
	// b = time.Time(t).AppendFormat(b, time.RFC3339)
	// b = append(b, '"')
	// return b, nil
	if y := time.Time(t).Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}

	return []byte(strconv.FormatInt(int64(time.Time(t).Unix()), 10)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *Time) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	// // Fractional seconds are handled implicitly by Parse.
	// var (
	// 	err error
	// 	tt  time.Time
	// )

	// tt, err = time.Parse(`"`+time.RFC3339+`"`, string(data))
	// *t = Time(tt)
	// Fractional seconds are handled implicitly by Parse.
	var (
		err error
		i   int64
	)
	i, err = strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	tt := time.Unix(i, 0)
	*t = Time(tt)
	return err
}

// MarshalText implements the encoding.TextMarshaler interface.
// The time is formatted in RFC 3339 format, with sub-second precision added if present.
func (t Time) MarshalText() ([]byte, error) {
	if y := time.Time(t).Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalText: year outside of range [0,9999]")
	}

	return []byte(strconv.FormatInt(int64(time.Time(t).Unix()), 10)), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The time is expected to be in RFC 3339 format.
func (t *Time) UnmarshalText(data []byte) error {
	// Fractional seconds are handled implicitly by Parse.
	var (
		err error
		i   int64
	)
	i, err = strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	tt := time.Unix(i, 0)
	*t = Time(tt)
	return err
}

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (t *Time) Scan(value interface{}) error {
	bytes, ok := value.(time.Time)
	if ok {
		*t = Time(bytes)
		return nil
	}
	return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (t Time) Value() (driver.Value, error) {
	var zeroTime time.Time
	if time.Time(t).UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return time.Time(t), nil
}
