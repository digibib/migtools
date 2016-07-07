// cleanitems cleans items marcdatabases for Nydalen|Bjørnholt-lærmidler.
//
// It performs the following operations:
//  * removes any due dates (952$q)
//  * removes items makred as lost, paid for ++
//    TODO get specifics! 952$7=? 952$1=?
//
// The result is dumped to standard out
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/boutros/marc"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("cleanitems: ")
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: cleanitems <marcdatabase>\n")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Detect format
	sniff := make([]byte, 64)
	_, err = f.Read(sniff)
	if err != nil {
		log.Fatal(err)
	}
	format := marc.DetectFormat(sniff)
	switch format {
	case marc.MARC, marc.LineMARC, marc.MARCXML:
		break
	default:
		log.Fatal("Unknown MARC format")
	}

	// rewind reader
	_, err = f.Seek(0, 0)
	if err != nil {
		log.Fatal(err)
	}

	dec := marc.NewDecoder(f, format)

	for rec, err := dec.Decode(); err != io.EOF; rec, err = dec.Decode() {
		if err != nil {
			log.Fatal(err)
		}
		stripDueDate(&rec)
		rec.DumpTo(os.Stdout, true)
	}
}

func stripDueDate(rec *marc.Record) {
	for i, f := range rec.DataFields {
		if f.Tag == "952" {
			for j, sf := range f.SubFields {
				if sf.Code == "q" {
					rec.DataFields[i].SubFields = append(rec.DataFields[i].SubFields[:j], rec.DataFields[i].SubFields[j+1:]...)
				}
			}
		}
	}
}
