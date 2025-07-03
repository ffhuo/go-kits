package uuid

import (
	"crypto/rand"
	"encoding"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

// ErrInvalidHex indicates that a hex string cannot be converted to an UUID.
var ErrInvalidHex = errors.New("the provided hex string is not a valid UUID")

// UUID is the BSON UUID type.
type UUID [12]byte

// NilUUID is the zero value for UUID.
var NilUUID UUID

var uuidCounter = readRandomUint32()
var processUnique = processUniqueBytes()

var _ encoding.TextMarshaler = UUID{}
var _ encoding.TextUnmarshaler = &UUID{}

// NewUUID generates a new UUID.
func NewUUID() UUID {
	return NewUUIDFromTimestamp(time.Now())
}

// NewUUIDFromTimestamp generates a new UUID based on the given time.
func NewUUIDFromTimestamp(timestamp time.Time) UUID {
	var b [12]byte

	binary.BigEndian.PutUint32(b[0:4], uint32(timestamp.Unix()))
	copy(b[4:9], processUnique[:])
	putUint24(b[9:12], atomic.AddUint32(&uuidCounter, 1))

	return b
}

// Timestamp extracts the time part of the ObjectId.
func (id UUID) Timestamp() time.Time {
	unixSecs := binary.BigEndian.Uint32(id[0:4])
	return time.Unix(int64(unixSecs), 0).UTC()
}

// Hex returns the hex encoding of the UUID as a string.
func (id UUID) Hex() string {
	var buf [24]byte
	hex.Encode(buf[:], id[:])
	return string(buf[:])
}

func (id UUID) String() string {
	return fmt.Sprintf("UUID(%q)", id.Hex())
}

// IsZero returns true if id is the empty UUID.
func (id UUID) IsZero() bool {
	return id == NilUUID
}

// UUIDFromHex creates a new UUID from a hex string. It returns an error if the hex string is not a
// valid UUID.
func UUIDFromHex(s string) (UUID, error) {
	if len(s) != 24 {
		return NilUUID, ErrInvalidHex
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return NilUUID, err
	}

	var oid [12]byte
	copy(oid[:], b)

	return oid, nil
}

// IsValidUUID returns true if the provided hex string represents a valid UUID and false if not.
func IsValidUUID(s string) bool {
	_, err := UUIDFromHex(s)
	return err == nil
}

// MarshalText returns the UUID as UTF-8-encoded text. Implementing this allows us to use UUID
// as a map key when marshalling JSON. See https://pkg.go.dev/encoding#TextMarshaler
func (id UUID) MarshalText() ([]byte, error) {
	return []byte(id.Hex()), nil
}

// UnmarshalText populates the byte slice with the UUID. Implementing this allows us to use UUID
// as a map key when unmarshalling JSON. See https://pkg.go.dev/encoding#TextUnmarshaler
func (id *UUID) UnmarshalText(b []byte) error {
	oid, err := UUIDFromHex(string(b))
	if err != nil {
		return err
	}
	*id = oid
	return nil
}

// MarshalJSON returns the UUID as a string
func (id UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.Hex())
}

// UnmarshalJSON populates the byte slice with the UUID. If the byte slice is 24 bytes long, it
// will be populated with the hex representation of the UUID. If the byte slice is twelve bytes
// long, it will be populated with the BSON representation of the UUID. This method also accepts empty strings and
// decodes them as NilUUID. For any other inputs, an error will be returned.
func (id *UUID) UnmarshalJSON(b []byte) error {
	// Ignore "null" to keep parity with the standard library. Decoding a JSON null into a non-pointer UUID field
	// will leave the field unchanged. For pointer values, encoding/json will set the pointer to nil and will not
	// enter the UnmarshalJSON hook.
	if string(b) == "null" {
		return nil
	}

	var err error
	switch len(b) {
	case 12:
		copy(id[:], b)
	default:
		// Extended JSON
		var res interface{}
		err := json.Unmarshal(b, &res)
		if err != nil {
			return err
		}
		str, ok := res.(string)
		if !ok {
			m, ok := res.(map[string]interface{})
			if !ok {
				return errors.New("not an extended JSON UUID")
			}
			oid, ok := m["$oid"]
			if !ok {
				return errors.New("not an extended JSON UUID")
			}
			str, ok = oid.(string)
			if !ok {
				return errors.New("not an extended JSON UUID")
			}
		}

		// An empty string is not a valid UUID, but we treat it as a special value that decodes as NilUUID.
		if len(str) == 0 {
			copy(id[:], NilUUID[:])
			return nil
		}

		if len(str) != 24 {
			return fmt.Errorf("cannot unmarshal into an UUID, the length must be 24 but it is %d", len(str))
		}

		_, err = hex.Decode(id[:], []byte(str))
		if err != nil {
			return err
		}
	}

	return err
}

func processUniqueBytes() [5]byte {
	var b [5]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot initialize objectid package with crypto.rand.Reader: %v", err))
	}

	return b
}

func readRandomUint32() uint32 {
	var b [4]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot initialize objectid package with crypto.rand.Reader: %v", err))
	}

	return (uint32(b[0]) << 0) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24)
}

func putUint24(b []byte, v uint32) {
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}
