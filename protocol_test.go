package main

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	type fields struct {
		bf *bytes.Buffer
	}
	tests := []struct {
		name    string
		fields  fields
		want    *Input
		wantErr bool
	}{
		{name: "array", fields: fields{
			bf: bytes.NewBuffer([]byte("*2\r\n$3\r\nbar\r\n$5\r\nhello\r\n")),
		}, want: &Input{
			Value: nil,
			Type:  ArrayType,
			Array: []*Input{
				&Input{
					Type:  BulkStringType,
					Value: []byte("bar"),
				},
				&Input{
					Type:  BulkStringType,
					Value: []byte("hello"),
				},
			},
		}},
		{name: "ok", fields: fields{
			bf: bytes.NewBuffer([]byte("+OK\r\n")),
		}, want: &Input{
			Value: []byte("OK"),
			Type:  StringType,
		}},
		{name: "1000", fields: fields{
			bf: bytes.NewBuffer([]byte(":1000\r\n")),
		}, want: &Input{
			Value: []byte("1000"),
			Type:  IntType,
		}},
		{name: "hello", fields: fields{
			bf: bytes.NewBuffer([]byte("$5\r\nhello\r\n")),
		}, want: &Input{
			Value: []byte("hello"),
			Type:  BulkStringType,
		}},
		{name: "less bulk string", fields: fields{
			bf: bytes.NewBuffer([]byte("$5\r\nhell\r\n")),
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				bf: tt.fields.bf,
			}
			got, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
