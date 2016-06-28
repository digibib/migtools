package main

import (
	"compress/gzip"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	imagesFile := flag.String("i", "", "input file (gzipped csv with columns: tnr, source)")
	host := flag.String("h", "", "host (namespace)")
	flag.Parse()
	if *imagesFile == "" || *host == "" {
		flag.Usage()
		os.Exit(1)
	}
	f, err := os.Open(*imagesFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}
	defer zr.Close()

	r := csv.NewReader(zr)
	fmt.Printf("PREFIX : <http://%s/ontology#>\n", *host)
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		url := fmt.Sprintf("http://static.deichman.no/bilder/%s/%s/1_thumb.jpg", rec[0], rec[1])
		fmt.Println("DELETE { ?pub :hasImage ?url }")
		fmt.Printf("INSERT { ?pub :hasImage %q . }\n", url)
		fmt.Printf("WHERE { ?pub :bibliofilPublicationID %q OPTIONAL { ?pub :hasImage ?url } }\n;\n", rec[0])
	}
}
