// patronmassage massages patron data for import into Koha
//
// input:
//   laaner: database export from Bibliofil
//   lmarc: database export from Bibliofil
//   lnel: database export from Bibliofil
//
// output:
//   patrons.csv:      patrons to be imported into Koha MySQL borrowers table
//   categories.sql    patron categories to be inserted into MySQL
//   branches.sql      branches to be inserted into MySQL

package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/boutros/marc"
)

var outDir *string

type Main struct {
	laanerIn, lmarcIn, lnelIn io.Reader
	laaner, lnel              map[int]map[string]string
	lmarc                     map[int]marc.Record
	numWorkers                int
}

func newMain(laaner, lmarc, lnel io.Reader, nw int) Main {
	return Main{
		laanerIn:   laaner,
		lmarcIn:    lmarc,
		lnelIn:     lnel,
		laaner:     make(map[int]map[string]string),
		lnel:       make(map[int]map[string]string),
		lmarc:      make(map[int]marc.Record),
		numWorkers: nw,
	}
}

func (m Main) indexLmarc(wg *sync.WaitGroup) {
	dec := marc.NewDecoder(m.lmarcIn, marc.LineMARC)
	for rec, err := dec.Decode(); err != io.EOF; rec, err = dec.Decode() {
		if err != nil {
			log.Fatal(err)
			// TODO continue?
		}
		n, err := borrowernumber(&rec)
		if err != nil {
			log.Println(err)
			rec.DumpTo(os.Stderr, true)
			continue
		}
		m.lmarc[n] = rec
	}
	wg.Done()
	log.Println("done indexing lmarc")
}

func (m Main) indexLaaner(wg *sync.WaitGroup) {
	dec := NewKVDecoder(m.laanerIn)
	for rec, err := dec.Decode(); err != io.EOF; rec, err = dec.Decode() {
		if err != nil {
			log.Fatal(err)
			// TODO continue?
		}
		if rec["ln_nr"] == "" {
			continue
		}
		n, err := strconv.Atoi(rec["ln_nr"])
		if err != nil {
			log.Fatal(err)
		}
		m.laaner[n] = rec
	}
	log.Println("done indexing laaner")
	wg.Done()
}

func (m Main) indexLnel(wg *sync.WaitGroup) {
	dec := NewKVDecoder(m.lnelIn)
	for rec, err := dec.Decode(); err != io.EOF; rec, err = dec.Decode() {
		if err != nil {
			log.Fatal(err)
			// TODO continue?
		}
		if rec["lnel_nr"] == "" {
			continue
		}
		n, err := strconv.Atoi(rec["lnel_nr"])
		if err != nil {
			log.Fatal(err)
		}
		m.lnel[n] = rec
	}
	log.Println("done indexing lnel")
	wg.Done()
}

func (m Main) Run() {
	log.Println("start indexing resources")

	var wg sync.WaitGroup
	wg.Add(3)
	go m.indexLmarc(&wg)
	go m.indexLaaner(&wg)
	go m.indexLnel(&wg)
	wg.Wait()

	log.Println("done indexing resources")

	jobs := make(chan int)
	patrons := make(chan patron)
	patronsF := mustCreate(filepath.Join(*outDir, "patrons.csv"))
	defer patronsF.Close()
	enc := csv.NewWriter(patronsF)
	defer enc.Flush()
	go func() {
		for p := range patrons {
			if err := enc.Write(patronCSVRow(p)); err != nil {
				log.Fatal(err)
			}
			wg.Done()
		}
	}()
	wg.Add(m.numWorkers)
	for i := 0; i < m.numWorkers; i++ {
		go func() {
			for lnr := range jobs {
				p := merge(m.lmarc[lnr], m.laaner[lnr], m.lnel[lnr])
				if !strings.HasPrefix(p.surname, "!!") {
					// deleted patrons are prefixed with !!
					wg.Add(1)
					patrons <- p
				}
			}
			wg.Done()
		}()
	}
	for lnr, _ := range m.laaner {
		jobs <- lnr
	}
	close(jobs)
	wg.Wait()
	close(patrons)
}

func main() {
	laaner := flag.String("laaner", "", "laaner dump")
	lmarc := flag.String("lmarc", "", "lmarc dump")
	lnel := flag.String("lnel", "", "lnel dump")
	numWorkers := flag.Int("n", 8, "number of concurrent workers")
	outDir = flag.String("outdir", "", "output directory (default to current working directory)")

	flag.Parse()

	if *laaner == "" || *lmarc == "" || *lnel == "" {
		flag.Usage()
		os.Exit(1)
	}

	laanerF := mustOpen(*laaner)
	defer laanerF.Close()
	lmarcF := mustOpen(*lmarc)
	defer lmarcF.Close()
	lnelF := mustOpen(*lnel)
	defer lnelF.Close()

	newMain(laanerF, lmarcF, lnelF, *numWorkers).Run()
}

func mustOpen(s string) *os.File {
	f, err := os.Open(s)
	if err != nil {
		panic(err)
	}
	return f
}

func mustCreate(s string) *os.File {
	f, err := os.Create(s)
	if err != nil {
		panic(err)
	}
	return f
}

func borrowernumber(r *marc.Record) (int, error) {
	for _, cf := range r.CtrlFields {
		if cf.Tag == "001" {
			return strconv.Atoi(cf.Value)
		}
	}
	return 0, errors.New("no borrowernumber in lmarc record")
}

func patronCSVRow(p patron) []string {
	row := make([]string, 10)
	row[0] = p.userid
	row[1] = p.cardnumber
	row[2] = p.surname
	row[3] = p.firstname
	row[4] = p.address
	row[5] = p.zipcode
	row[6] = p.city
	row[7] = p.country
	row[8] = p.smsalertnumber
	row[9] = p.email
	return row
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("patronmassage: ")
}
