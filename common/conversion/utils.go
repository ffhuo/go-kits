package conversion

func GetBytes(v interface{}) (b []byte, ok bool) {
	switch d := v.(type) {
	case []byte:
		return d, true
	case string:
		return StringToBytes(d), ok
	}
	return nil, false
}
