// cleanitems cleans items marcdatabases for Nydalen|Bjørnholt-lærmidler.
//
// It performs the following operations:
//  * removes any due dates (952$q)
//  * removes items makred as lost, paid for ++
//    TODO get specifics! 952$7=? 952$1=?
//  * sets item type 952$y to "L" for "Læremidler"
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
	enc := marc.NewEncoder(os.Stdout, format)
	for rec, err := dec.Decode(); err != io.EOF; rec, err = dec.Decode() {
		if err != nil {
			log.Fatal(err)
		}
		stripDueDate(&rec)
		removeItems(&rec)
		stripStatusCodes(&rec)
		setItemType(&rec)
		if err := enc.Encode(rec); err != nil {
			log.Fatal(err)
		}
	}
}

// stripDueDateremove any occurency of 952$q from MARC record.
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

// removeItems removes any items marked as: "Tapt", "Regning", "Tapt. Regning betalt", "påstått ikke lånt;
// that is, where 952$1 equals "1", "12", "8", "5"
func removeItems(rec *marc.Record) {
	for i, f := range rec.DataFields {
		if f.Tag == "952" {
			for _, sf := range f.SubFields {
				if sf.Code == "1" {
					switch sf.Value {
					case "1", "12", "8", "5":
						rec.DataFields = append(rec.DataFields[:i], rec.DataFields[min(i+1, len(rec.DataFields)):]...)
						removeItems(rec) // need to start over, since we mutated the slice we're ranging over
					default:
					}
				}
			}
		}
	}
}

func setItemType(rec *marc.Record) {
	for i, f := range rec.DataFields {
		if f.Tag == "952" {
			for j, sf := range f.SubFields {
				if sf.Code == "y" {
					rec.DataFields[i].SubFields[j].Value = "L"
				}
			}
		}
	}
}

// stripStatusCodes remove some status codes from 952$7 and 952$1, namely:
//   lost: "på vidvanke", "return eieravdeling", "til henteavdeling"
//   notloan: "ny"
func stripStatusCodes(rec *marc.Record) {
	for i, f := range rec.DataFields {
		if f.Tag == "952" {
			for j, sf := range f.SubFields {
				if sf.Code == "1" && (sf.Value == "9" || sf.Value == "10" || sf.Value == "11") {
					// remove lost codes for "på vidvanke", "return eieravdeling", "til henteavdeling"
					rec.DataFields[i].SubFields = append(
						rec.DataFields[i].SubFields[:j],
						rec.DataFields[i].SubFields[min(j+1, len(rec.DataFields[i].SubFields)):]...)
					stripStatusCodes(rec) // need to start over, since we mutated the slice we're ranging over
				}
				if sf.Code == "7" && (sf.Value == "2") {
					// remove notforloan codes for "ny"
					rec.DataFields[i].SubFields = append(
						rec.DataFields[i].SubFields[:j],
						rec.DataFields[i].SubFields[min(j+1, len(rec.DataFields[i].SubFields)):]...)
					stripStatusCodes(rec) // need to start over, since we mutated the slice we're ranging over
				}

			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
