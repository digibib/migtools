package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	idsFile := flag.String("i", "", "input file (tab-separated Koha-biblionumber and Bibliofil-titittelnr)")
	host := flag.String("h", "", "host (namespace)")
	flag.Parse()
	if *idsFile == "" || *host == "" {
		flag.Usage()
		os.Exit(1)
	}
	f, err := os.Open(*idsFile)
	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(f)
	r.Comma = '\t'
	fmt.Printf("PREFIX : <http://%s/ontology#>\n", *host)
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("INSERT { ?pub :recordId %q }\n", rec[0])
		fmt.Printf("WHERE { ?pub :bibliofilPublicationID %q }; \n", trimLeadingZeroes(rec[1]))
	}
}

func trimLeadingZeroes(s string) string {
	for i, r := range s {
		if r != '0' {
			return s[i:]
		}
	}
	return s
}
