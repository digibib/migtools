package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/knakk/rdf"
)

func main() {
	input := flag.String("f", "", "input file (N-triples")
	flag.Parse()
	if *input == "" {
		flag.PrintDefaults()
	}

	f, err := os.Open(*input)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	dec := rdf.NewTripleDecoder(f, rdf.Turtle)
	bnodes := make(map[string][]rdf.Triple)
	for tr, err := dec.Decode(); err != io.EOF; tr, err = dec.Decode() {
		if err != nil {
			log.Fatal(err)
		}
		if tr.Subj.Type() == rdf.TermBlank {
			bnodes[tr.Subj.String()] = append(bnodes[tr.Subj.String()], tr)
		}
	}

	if _, err := f.Seek(0, 0); err != nil {
		log.Fatal(err)
	}

	dec = rdf.NewTripleDecoder(f, rdf.Turtle)
	for tr, err := dec.Decode(); err != io.EOF; tr, err = dec.Decode() {
		if err != nil {
			log.Fatal(err)
		}
		if tr.Obj.Type() == rdf.TermBlank {
			fmt.Printf("%s %s [ ", tr.Subj.Serialize(rdf.NTriples), tr.Pred.Serialize(rdf.NTriples))
			l := len(bnodes[tr.Obj.String()])
			for i, tr := range bnodes[tr.Obj.String()] {
				if i > 0 && i < l {
					fmt.Print(" ; ")
				}
				fmt.Printf("%s %s", tr.Pred.Serialize(rdf.NTriples), tr.Obj.Serialize(rdf.NTriples))
			}
			fmt.Print(" . ]\n")
		} else if tr.Subj.Type() != rdf.TermBlank {
			fmt.Print(tr.Serialize(rdf.NTriples))
		}
	}

}
