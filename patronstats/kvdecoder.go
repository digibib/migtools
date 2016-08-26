package main

import (
	"bufio"
	"errors"
	"io"
	"unicode/utf8"
)

var errEndOfRecord = errors.New("^")

const eof = rune(-1)

type KVDecoder struct {
	r     *bufio.Reader
	line  []byte // line beeing scanned
	start int    // pos of current token
	pos   int    // byte position in line
}

func NewKVDecoder(r io.Reader) *KVDecoder {
	return &KVDecoder{
		r: bufio.NewReader(r),
	}
}

func (d *KVDecoder) next() rune {
	if d.pos == len(d.line) {
		line, err := d.r.ReadBytes('\n')
		if err != nil && len(line) == 0 {
			return eof
		}
		d.line = line
		d.start = 0
		d.pos = 0
	}

	r, w := utf8.DecodeRune(d.line[d.pos:])
	d.pos += w

	return r
}

func (d *KVDecoder) peek() rune {
	r, _ := utf8.DecodeRune(d.line[d.pos:])
	return r
}

func (d *KVDecoder) decodeKey() (string, error) {
	d.start = d.pos
	for r := d.next(); r != ' '; r = d.next() {
		if r == eof {
			return "", io.EOF
		}
		if r == '^' {
			return "", errEndOfRecord
		}
	}
	return string(d.line[d.start : d.pos-1]), nil
}

func (d *KVDecoder) decodeVal() (string, error) {
	d.pos++ // scan |
	d.start = d.pos
again:
	for r := d.next(); r != '|'; r = d.next() {
		if r == '\n' {
			// bad input data; got EOL before '|'
			// keep track of token and consume another line
			tok1 := string(d.line[d.start : d.pos-1])
			d.pos--
			tok2, err := d.decodeVal()
			if err != nil {
				return "", err
			}
			return tok1 + tok2, nil
		}
		if r == eof {
			return "", io.EOF
		}
		if r == '^' {
			return "", errEndOfRecord
		}
	}
	if d.peek() != '\n' {
		// bad input data; got pipe in value
		// keep on consuming until EOL
		d.pos++
		goto again
	}

	return string(d.line[d.start : d.pos-1]), nil
}

func (d *KVDecoder) Decode() (map[string]string, error) {
	res := make(map[string]string)

parseRecord:
	for {
		k, err := d.decodeKey()
		switch err {
		case io.EOF:
			if len(res) > 0 {
				// we have a record with data, leave io.EOF for next call to Deocde()
				return res, nil
			}
			return nil, io.EOF
		case errEndOfRecord:
			break parseRecord
		case nil:
			// ok
		default:
			return nil, err
		}

		v, err := d.decodeVal()
		switch err {
		case io.EOF:
			if len(res) > 0 {
				// we have a record with data, leave io.EOF for next call to Deocde()
				return res, nil
			}
			return nil, io.EOF
		case errEndOfRecord:
			break parseRecord
		case nil:
			// ok
		default:
			return nil, err
		}

		res[k] = v
	}
	return res, nil
}
