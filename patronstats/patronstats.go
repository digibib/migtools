package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/boutros/marc"
)

type Main struct {
	laanerIn, lmarcIn, lnelIn io.Reader
	laaner, lnel              map[int]map[string]string
	lmarc                     map[int]*marc.Record
	numWorkers                int
	branches                  map[string]string
}

func newMain(laaner, lmarc, lnel io.Reader, nw int) *Main {
	return &Main{
		laanerIn:   laaner,
		lmarcIn:    lmarc,
		lnelIn:     lnel,
		laaner:     make(map[int]map[string]string),
		lnel:       make(map[int]map[string]string),
		lmarc:      make(map[int]*marc.Record),
		numWorkers: nw,
		branches:   make(map[string]string),
	}
}

func (m *Main) indexLmarc(wg *sync.WaitGroup) {
	dec := marc.NewDecoder(m.lmarcIn, marc.LineMARC)
	for rec, err := dec.Decode(); err != io.EOF; rec, err = dec.Decode() {
		if err != nil {
			log.Fatal(err)
			// TODO continue?
		}
		n, err := borrowernumber(rec)
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

func (m *Main) indexLaaner(wg *sync.WaitGroup) {
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

func (m *Main) indexLnel(wg *sync.WaitGroup) {
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

func (m *Main) Run() {
	log.Println("start indexing resources")

	var wg sync.WaitGroup
	wg.Add(3)
	go m.indexLmarc(&wg)
	go m.indexLaaner(&wg)
	go m.indexLnel(&wg)
	wg.Wait()

	log.Println("done indexing resources")

	patrons := make([]patron, 0, 200000)

	for i, _ := range m.laaner {
		p := merge(m.lmarc[i], m.laaner[i], m.lnel[i])

		if !strings.HasPrefix(p.surname, "!!") {
			// deleted patrons are prefixed with !!
			if p.cardnumber == "" {
				p.cardnumber = p.userid
			}
			patrons = append(patrons, p)
		}
	}

	enc := csv.NewWriter(os.Stdout)
	defer enc.Flush()

	for _, sel := range selections {
		c := 0
		for _, p := range patrons {
			if sel.IncludeFn(p) {
				/*if err := enc.Write(patronCSVRow(p)); err != nil {
					log.Fatal(err)
				}*/
				c++
			}
		}
		fmt.Printf("\n%s: %d\n", sel.Desc, c)
	}
}

func main() {
	var (
		laaner     = flag.String("laaner", "/home/boutros/src/github.com/digibib/ls.ext/migration/data/data.laaner.20160819-073100.txt", "laaner dump")
		lmarc      = flag.String("lmarc", "/home/boutros/src/github.com/digibib/ls.ext/migration/data/data.lmarc.20160819-073115.txt", "lmarc dump")
		lnel       = flag.String("lnel", "/home/boutros/src/github.com/digibib/ls.ext/migration/data/data.lnel.20160819-073113.txt", "lnel dump")
		numWorkers = flag.Int("n", 8, "number of concurrent workers")
	)

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

	m := newMain(laanerF, lmarcF, lnelF, *numWorkers)
	m.Run()
}

func mustOpen(s string) *os.File {
	f, err := os.Open(s)
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
	row := make([]string, 3)
	row[0] = p.surname
	row[1] = p.firstname
	row[2] = p.email
	return row
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("patronmassage: ")
}

var selections = []struct {
	Desc      string
	IncludeFn func(patron) bool
}{
	{
		"Aktive lånere (lånt i 2015 el 2016) med epostadresse",
		func(p patron) bool { return p.email != "" && isActive(p.TEMP_sistelaan) },
	},
	{
		"Lånere med epostadresse som har lånt de siste 12 månedene",
		func(p patron) bool { return p.email != "" && isActiveLastNMonths(p.TEMP_sistelaan, 12) },
	},
	{
		"Lånere med epostadresse som har lånt de siste 24 månedene",
		func(p patron) bool { return p.email != "" && isActiveLastNMonths(p.TEMP_sistelaan, 24) },
	},
	{
		"Lånere med epostadresse som har lånt de siste 36 månedene",
		func(p patron) bool { return p.email != "" && isActiveLastNMonths(p.TEMP_sistelaan, 36) },
	},
	{
		"Lånere med epostadresse som har brukt huskeliste",
		func(p patron) bool { return p.email != "" && p.TEMP_huskeliste },
	},
	{
		"Lånere med epostadresse som har lagret historikk",
		func(p patron) bool { return p.email != "" && p.privacy == 0 },
	},
	{
		"Lånere med epostadresse som har brukt famileMappaMi",
		func(p patron) bool { return p.email != "" && p.TEMP_familie },
	},
	{
		"Lånere med epostadresse som har brukt interesseområder",
		func(p patron) bool { return p.email != "" && p.TEMP_interesse },
	},
	{
		"Lånere med epostadresse",
		func(p patron) bool { return p.email != "" },
	},
}

func isActive(s string) bool {
	// Date format: 2016-03-07
	return len(s) == 10 && (s[:4] == "2015" || s[:4] == "2016")
}

func isActiveLastNMonths(s string, n int) bool {
	// Date format: 2016-03-07
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return false
	}
	return len(s) == 10 && int((time.Since(t).Hours()/730.485)) <= n // TODO verify 730.485 calc
}
