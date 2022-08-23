package encode

import (
	"bytes"
	"testing"
)

func TestJSONEncode_Encode(t *testing.T) {
	type fields map[string]interface{}
	tests := []struct {
		name    string
		fields  fields
		wantW   string
		wantErr bool
	}{
		{
			name: "test_encode",
			fields: fields{
				"name": "test1",
				"data": 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := NewJSONEncoder(tt.fields)
			w := &bytes.Buffer{}
			if err := j.Encode(w); (err != nil) != tt.wantErr {
				t.Errorf("JSONEncode.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("JSONEncode.Encode() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
