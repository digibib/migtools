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
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/boutros/marc"
)

type Main struct {
	laanerIn, lmarcIn, lnelIn io.Reader
	laaner, lnel              map[int]map[string]string
	lmarc                     map[int]marc.Record
	numWorkers                int
}

func newMain(laaner, lmarc, lnel io.Reader) Main {
	return Main{
		laanerIn: laaner,
		lmarcIn:  lmarc,
		lnelIn:   lnel,
		laaner:   make(map[int]map[string]string),
		lnel:     make(map[int]map[string]string),
		lmarc:    make(map[int]marc.Record),
	}
}

func (m Main) Run() error {
	log.Println("start indexing resources")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
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
		log.Println("done indexing lmarc")
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		// TODO
		log.Println("done indexing laaner")
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		// TODO
		log.Println("done indexing lnel")
		wg.Done()
	}()

	wg.Wait()
	log.Println("done indexing resources")
	return nil
}

func main() {
	laaner := flag.String("laaner", "", "laaner dump")
	lmarc := flag.String("lmarc", "", "lmarc dump")
	lnel := flag.String("lnel", "", "lnel dump")
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

	m := newMain(laanerF, lmarcF, lnelF)
	if err := m.Run(); err != nil {
		log.Fatal(err)
	}
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

func init() {
	log.SetFlags(0)
	log.SetPrefix("patronmassage: ")
}
