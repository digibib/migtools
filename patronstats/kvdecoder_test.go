package main

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

const input = `a |12345|
b |æøå|
c |bad data|or?|
xyz_fieldname |
1999|
^
a |1|
b |."a"|
^
a |2|
b ||
^`

func TestKVDecoder(t *testing.T) {
	dec := NewKVDecoder(bytes.NewBufferString(input))

	wants := [...]map[string]string{
		map[string]string{
			"a":             "12345",
			"b":             "æøå",
			"c":             "bad data|or?",
			"xyz_fieldname": "1999",
		},
		map[string]string{
			"a": "1",
			"b": `."a"`,
		},
		map[string]string{
			"a": "2",
			"b": "",
		},
	}
	for i := 0; i < len(wants); i++ {
		rec, err := dec.Decode()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(rec, wants[i]) {
			t.Fatalf("got %v; want %v", rec, wants[i])
		}
	}
	_, err := dec.Decode()
	if err != io.EOF {
		t.Fatalf("got %v; want io.EOF to signify end of stream", err)
	}
}
