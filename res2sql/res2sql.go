package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"text/template"
	"time"
)

const (
	noDateFormat    = "02/01/2006"
	mysqlDateFormat = "2006-01-02"
	sqlTmpl         = `
INSERT INTO reserves
  (borrowernumber, reservedate, biblionumber, branchcode, priority, found, expirationdate)
SELECT borrowers.borrowernumber,
       '{{.ReserveDate}}',
       '{{.Biblionumber}}',
       '{{.Branchcode}}',
       '{{.Priority}}',
       '{{.Status}}',
       '{{.ExpirationDate}}'
FROM borrowers WHERE borrowers.userid='{{.Borrowernumber}}';
`

	sqlTmplWithEx = `
INSERT INTO reserves
  (borrowernumber, reservedate, biblionumber, branchcode, priority, found, itemnumber, expirationdate)
SELECT borrowers.borrowernumber,
       '{{.ReserveDate}}',
       '{{.Biblionumber}}',
       '{{.Branchcode}}',
       '{{.Priority}}',
       '{{.Status}}',
       items.itemnumber,
       '{{.ExpirationDate}}'
FROM borrowers JOIN items
WHERE barcode='{{.Barcode}}' AND borrowers.userid='{{.Borrowernumber}}';
`
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("res2sql: ")
}

type Reserve struct {
	Borrowernumber string
	Biblionumber   string
	Priority       string
	Exnr           string
	Status         string
	ReserveDate    string
	ExpirationDate string
	Branchcode     string
	Barcode        string
}

type Reserves []Reserve

func (r Reserves) Len() int           { return len(r) }
func (r Reserves) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Reserves) Less(i, j int) bool { return mustInt(r[i].Priority) < mustInt(r[j].Priority) }

func main() {
	resInput := flag.String("res", "", "res dump")
	flag.Parse()

	if *resInput == "" {
		flag.PrintDefaults()
	}

	f, err := os.Open(*resInput)
	if err != nil {
		log.Fatal(err)
	}

	dec := NewKVDecoder(f)

	noexTmpl := template.Must(template.New("noex").Parse(sqlTmpl))
	exTmpl := template.Must(template.New("ex").Parse(sqlTmplWithEx))

	all := make(map[string]Reserves) // map[biblionumber]reserves
	for rec, err := dec.Decode(); err != io.EOF; rec, err = dec.Decode() {
		if err != nil {
			log.Fatal(err)
		}

		if rec["res_exnr"] == "998" {
			// eksemplarnr 998 = innlån. Hopper over disse
			continue
		}

		res := Reserve{
			Biblionumber:   rec["res_titnr"],
			Priority:       rec["res_koenr"],
			Exnr:           rec["res_exnr"],
			Branchcode:     rec["res_hentavd"],
			Borrowernumber: rec["res_laanr"],
		}

		if res.Biblionumber == "" {
			log.Println("missing biblionumber")
			log.Printf("skipping record: %+v", rec)
			continue
		}

		// status
		switch rec["res_stat"] {
		case "i":
			res.Status = "W" // hentehylle
		case "y":
			res.Status = "T" // på vei til henteavdeling
		}

		// reservedate
		d, err := time.Parse(noDateFormat, rec["res_dat"])
		if err != nil {
			log.Fatal(err) // TODO continue?
		}
		res.ReserveDate = d.Format(mysqlDateFormat)

		// exiprationdate
		if rec["res_forfall"] != "00/00/0000" {
			d, err := time.Parse(noDateFormat, rec["res_forfall"])
			if err != nil {
				log.Println(err)
				log.Printf("skipping record: %+v", rec)
				continue
			}
			res.ExpirationDate = d.Format(mysqlDateFormat)
		}

		// generate barcode where specific item is reserved
		if res.Exnr != "0" {
			res.Barcode = fmt.Sprintf("0301%07d%03d", mustInt(res.Biblionumber), mustInt(res.Exnr))
		}

		all[res.Biblionumber] = append(all[res.Biblionumber], res)
	}

	for biblionr, _ := range all {
		sort.Sort(all[biblionr])
		for i, res := range all[biblionr] {
			res.Priority = strconv.Itoa(i + 1)
			t := noexTmpl // biblio-level reserve
			if res.Barcode != "" {
				// specific copy is reserved
				t = exTmpl
			}
			if err := t.Execute(os.Stdout, res); err != nil {
				log.Fatal(err)
			}
		}
	}

}

func mustInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}
