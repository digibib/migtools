package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	branchFile := flag.String("i", "", "input file (tab-separated biblionr and branches)")
	host := flag.String("h", "", "host (namespace)")
	flag.Parse()
	if *branchFile == "" || *host == "" {
		flag.Usage()
		os.Exit(1)
	}
	f, err := os.Open(*branchFile)
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

		fmt.Println("DELETE { ?pub :hasHoldingBranch ?branch }")
		fmt.Printf("INSERT { ?pub :hasHoldingBranch %s . }\n", quoteBranches(rec[1]))
		fmt.Printf("WHERE { ?pub :recordID %q OPTIONAL { ?pub :hasHoldingBranch ?branch } }\n;\n", rec[0])
	}
}

func quoteBranches(s string) string {
	branches := strings.Split(s, ",")
	for i := range branches {
		branches[i] = fmt.Sprintf("%q", branches[i])
	}
	return strings.Join(branches, ",")
}
