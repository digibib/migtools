// catmassage extracts, filters, merges and massages catalogue and items data.
//
// input:
//   vmarc: database export from Bibliofil
//   emarc: database export from Bibliofil
//
// output:
//   catalogue.mrc:      massaged catalogue with items information in MARC field 952, to be imported with bulcmarkimport
//   catalogue.marcxml:  massaged catalogue without item information, to be converted to RDF with migmarc2rdf
//   loans.csv:          active loans, to be inserted into MySQL after bulkmarcimport
//   bjorndalen.marcxml: catalogue with items belonging to "bjørndalen-læremidler"
//   nydalen.marcxml:    catalogue with items belonging to "nydalen-læremidler"
//   branches.sql:       holding branches extracted from items, to be inserted in MySQL before bulkmarcimport
//   itypes.sql          item types to be inserted in MySQL before bulkmarcimport
//   avalues.sql         authorized values (for status codes), to be inserted in MySQL before bulkmarcimport

package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/boutros/marc"
)

var (
	xmlHeader = []byte(`<?xml version="1.0" encoding="UTF-8"?><collection xmlns="http://www.loc.gov/MARC21/slim">`)
	xmlFooter = []byte(`</collection>`)

	prefixTitnr = []byte("ex_titnr |")

	rgx14days = regexp.MustCompile(`(?i)DG|ED|EE|EF|EG`)
)

// Main represents the main program execution
type Main struct {
	vmarc         io.Reader
	exemp         io.ReadSeeker
	outMerged     io.Writer
	outNoItems    io.Writer
	outLoans      io.Writer
	outBjorndalen io.Writer
	outNydalen    io.Writer
	limit         int
	skip          int
	branches      map[string]string
}

func newMain(vmarc io.Reader, exemp io.ReadSeeker, outMerged io.Writer, outNoItems io.Writer, limit int, skip int) *Main {
	return &Main{
		vmarc:      vmarc,
		exemp:      exemp,
		outMerged:  outMerged,
		outNoItems: outNoItems,
		limit:      limit,
		skip:       skip,
		branches:   make(map[string]string),
	}
}

func (m *Main) Run() error {

	// Create an index of the exemplar database by title number.
	// The DB is sorted by title number and copy number (ex_titnr and ex_exnr),
	// so all copies can be read sequentially.

	exemp := make(map[string]int64) // key=title, value=position
	n := 0                          // position in file
	scanner := bufio.NewScanner(m.exemp)
	c := 0
	for scanner.Scan() {
		if bytes.HasPrefix(scanner.Bytes(), prefixTitnr) {
			tnr := string(scanner.Bytes()[len(prefixTitnr) : len(scanner.Bytes())-1])
			if _, ok := exemp[tnr]; !ok {
				exemp[tnr] = int64(n)
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		n += len(scanner.Bytes()) + 1 // incl newline
	}

	laanF, err := os.Create("laan.csv")
	if err != nil {
		return err
	}
	defer laanF.Close()
	csv := csv.NewWriter(laanF)
	csv.Comma = '|'

	// Write XML header to catalogue.marcxml
	_, err = m.outNoItems.Write(xmlHeader)
	if err != nil {
		return err
	}

	// Loop over records in database, and merge exemplar info into field 952

	dec := marc.NewDecoder(m.vmarc, marc.LineMARC)
	encMARC := marc.NewEncoder(m.outMerged, marc.MARC)
	encMARCXML := marc.NewEncoder(m.outNoItems, marc.MARCXML)

	skipCount := 0
	if m.skip > 0 {
		log.Printf("Skipping first %d records\n", m.skip)
	}

	for r, err := dec.Decode(); err != io.EOF; r, err = dec.Decode() {
		if err != nil {
			return err
		}

		switch r.Leader[5:6] {
		case "f", "e", "i", "l", "t", "m", "d":
			// ignorer fjernlån/innlån/depot/slettede poster
			continue
		}

		if skipCount < m.skip {
			skipCount++
			continue
		}

		// write MARCXML record, before merging in items
		if err := encMARCXML.Encode(r); err != nil {
			return err
		}

		tnr := titleNumber(&r)
		tnrInt, err := strconv.Atoi(tnr)
		if err != nil {
			log.Println("Title number not an integer:", tnr)
			log.Println("See MARC record below (ignored):")
			r.DumpTo(os.Stderr, true)
			continue
		}

		// Add 942 field (default item type)
		v := firstVal(&r, "019", "b")
		if v == "" {
			v = "X"
		}
		r.DataFields = append(r.DataFields, marc.DField{
			Tag:       "942",
			SubFields: marc.SubFields{marc.SubField{Code: "y", Value: v}},
		})

		if pos, ok := exemp[tnr]; ok {
			// seek to first occurence of titlenumber in exemp database
			if _, err := m.exemp.Seek(pos, 0); err != nil {

				return err
			}

			scanner := bufio.NewScanner(m.exemp)
			f := marc.DField{Tag: "952"}
			var loanRow []string
			onLoan := false

			// parse exemplar information
		findExemplars:
			for scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return err
				}

				if i := bytes.Index(scanner.Bytes(), []byte(" ")); i != -1 {
					switch string(scanner.Bytes()[:i]) {
					case "ex_titnr":
						// check if we are still parsing exemplars of current title number
						if getValue(scanner.Bytes()) != tnr {
							break findExemplars
						}
						loanRow = make([]string, 6)
						loanRow[0] = tnr
					case "ex_exnr":
						// 952$t copy number
						f.SubFields = append(f.SubFields, marc.SubField{Code: "t", Value: getValue(scanner.Bytes())})
						// 952$p barcode - generated from titlenumber and barcode
						c, err := strconv.Atoi(getValue(scanner.Bytes()))
						if err != nil {
							log.Println("Title number: ", tnr, "Copy number not a number:", getValue(scanner.Bytes()))
							continue
						}
						barcode := fmt.Sprintf("0301%07d%03d", tnrInt, c)
						f.SubFields = append(f.SubFields, marc.SubField{Code: "p", Value: barcode})
						loanRow[1] = getValue(scanner.Bytes())
						loanRow[5] = barcode
					case "ex_avd":
						// 952$a branchcode and
						// 952$b holding branch (the same for now, possibly depot)
						bCode := getValue(scanner.Bytes())
						f.SubFields = append(f.SubFields, marc.SubField{Code: "a", Value: bCode})
						f.SubFields = append(f.SubFields, marc.SubField{Code: "b", Value: bCode})

						// Keep track of which branchcodes that are found
						if bLabel, ok := branchCodes[bCode]; ok {
							m.branches[bCode] = bLabel
						} else {
							m.branches[bCode] = "Missing label for branch: " + bCode
						}
					case "ex_plass":
						// 952$c shelving location (authorized value? TODO check)
						f.SubFields = append(f.SubFields, marc.SubField{Code: "c", Value: getValue(scanner.Bytes())})
					case "ex_hylle":
					case "ex_note":
						// 952$z public note
						if v := getValue(scanner.Bytes()); v != "" {
							f.SubFields = append(f.SubFields, marc.SubField{Code: "z", Value: v})
						}
					case "ex_bind":
						// 952$h volume and issue information, flerbindsverk?
						// Vises som "publication details" i grensesnittet. (Serienummererering/kronologi)
						if v := getValue(scanner.Bytes()); v != "0" {
							f.SubFields = append(f.SubFields, marc.SubField{Code: "h", Value: v})
						}
					case "ex_aar":
					case "ex_status":
						// Eksemplarstatus - mappes til autoriserte verdier i Koha.
						// Alle statuser er varianter av "Ikke til utlån", og "Tapt"
						v := getValue(scanner.Bytes())
						if sf, ok := statusCodes[v]; ok {
							f.SubFields = append(f.SubFields, sf)
						}
						if v == "u" {
							onLoan = true
						}
					case "ex_resstat":
					case "ex_laanstat":
						//952$m total renewals
						if v := getValue(scanner.Bytes()); v != "" {
							// Antall fornyelser som en char. Første fornyelse blir "1", andre "2" osv.
							// Dersom det fornyes over 9 ganger så blir det ":", ";", "<" osv. Følger ascii-tabellen.
							f.SubFields = append(f.SubFields, marc.SubField{
								Code:  "m",
								Value: strconv.FormatInt(int64(v[0]-48), 10),
							})
							loanRow[2] = v
						}
					case "ex_utlkode":
						if v := getValue(scanner.Bytes()); v == "e" || v == "r" {
							// autorisert verdi:
							// referanseverk: ikke til utlån
							f.SubFields = append(f.SubFields, marc.SubField{Code: "5", Value: "2"})
						}
					case "ex_laanr":
						loanRow[3] = strings.TrimPrefix(getValue(scanner.Bytes()), "-")
					case "ex_laantid":
					case "ex_forfall":
						//952$q due date (if checked out)
						if v := getValue(scanner.Bytes()); v != "00/00/0000" {
							if len(v) != 10 {
								log.Println("Unknown date format (ex_laantid):", v)
								break
							}
							forfall := fmt.Sprintf("%s-%s-%s", v[6:10], v[3:5], v[0:2])
							f.SubFields = append(f.SubFields, marc.SubField{
								Code:  "q",
								Value: forfall,
							})
							loanRow[4] = forfall
						}
					case "ex_purrdat":
					case "ex_antpurr":
					case "ex_etikett":
					case "ex_antlaan":
						// 952$l total checkouts
						f.SubFields = append(f.SubFields, marc.SubField{Code: "l", Value: getValue(scanner.Bytes())})
					case "ex_kl_sett":
					case "ex_strek":
					}
					continue
				}

				// End of record reached; append field to record unless it's empty
				if bytes.Equal(scanner.Bytes(), []byte("^")) && len(f.SubFields) > 0 {

					// 952$o full call number (hyllesignatur)
					// TODO factour out this string concatination
					callnumber := firstVal(&r, "090", "a")
					if v := firstVal(&r, "090", "b"); v != "" {
						if len(callnumber) > 0 {
							callnumber += " "
						}
						callnumber += v
					}
					if v := firstVal(&r, "090", "c"); v != "" {
						if len(callnumber) > 0 {
							callnumber += " "
						}
						callnumber += v
					}
					if v := firstVal(&r, "090", "d"); v != "" {
						if len(callnumber) > 0 {
							callnumber += " "
						}
						callnumber += v
					}
					if callnumber != "" {
						f.SubFields = append(f.SubFields, marc.SubField{Code: "o", Value: callnumber})
					}

					// Add item type (used for issuing rule) based on item type from record:
					recType := firstVal(&r, "942", "y")
					iType := "28" // default, 28 days checkout time
					if rgx14days.MatchString(recType) {
						iType = "14" // 14 days checkout time for some formats (CDs/DVDs)
					}
					// TODO ikke til utlån? eller skal de ligge i Not for loan-statuskodene?
					f.SubFields = append(f.SubFields, marc.SubField{Code: "y", Value: iType})

					r.DataFields = append(r.DataFields, f)
					f = marc.DField{Tag: "952"}

					if onLoan {
						// write CSV row to loan.csv
						err := csv.Write(loanRow)
						if err != nil {
							return err
						}
						onLoan = false
					}
				}
			}
		}

		if err != nil {
			fmt.Println(err)
			return err
		}
		if err = encMARC.Encode(r); err != nil {
			log.Println(err)
			log.Println("bibliofil titellnummer: ", titleNumber(&r))
			// TODO fail on IO errors:
			// return err
		}

		c++
		if c == m.limit {
			break
		}
	}

	encMARC.Flush()
	encMARCXML.Flush()
	csv.Flush()

	// write XML footer
	_, err = m.outNoItems.Write(xmlFooter)
	if err != nil {
		return err
	}

	return nil
}

// titleNumber returns the Record's title number from the 001 control field,
// stripping it of any leading zeros.
func titleNumber(r *marc.Record) string {
	for _, f := range r.CtrlFields {
		if f.Tag == "001" {
			i := 0
			for ; i < len(f.Value); i++ {
				if f.Value[i] != '0' {
					break
				}
			}
			return f.Value[i:]
		}
	}
	return ""
}

// firstVal returns the first value of a given tag and substring code
// of a Record, or empty string if not found.
func firstVal(r *marc.Record, tag string, code string) string {
	for _, f := range r.DataFields {
		if f.Tag == tag {
			for _, s := range f.SubFields {
				if s.Code == code {
					return s.Value
				}
			}
		}
	}
	return ""
}

// getValue returns the value from an line in exemplar database.
// Ex: []byte("ex_avd |hutl|") would return the string "hutl"
func getValue(b []byte) string {
	if i := bytes.Index(b, []byte("|")); i != -1 {
		if len(b) > i+1 {
			return string(b[i+1 : len(b)-1])
		}
	}
	return ""
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("catmassage: ")
}

func main() {
	var (
		vmarc  = flag.String("vmarc", "/home/boutros/src/github.com/digibib/ls.ext/migration/example_data/data.vmarc.20141020-084813.txt", "catalogue database in line-marc")
		exemp  = flag.String("exemp", "/home/boutros/src/github.com/digibib/ls.ext/migration/example_data/data.exemp.20141020-085129.txt", "exemplar database key-val")
		limit  = flag.Int("limit", -1, "stop after n records")
		skip   = flag.Int("skip", 0, "skip first n records")
		outDir = flag.String("outdir", "", "output directory (default to current working directory)")
	)

	flag.Parse()

	if *vmarc == "" || *exemp == "" {
		flag.Usage()
		os.Exit(1)
	}

	outMerged := mustCreate(filepath.Join(*outDir, "catalogue.mrc"))
	defer outMerged.Close()

	outNoItems := mustCreate(filepath.Join(*outDir, "catalogue.marcxml"))
	defer outNoItems.Close()

	vmarcF := mustOpen(*vmarc)
	defer vmarcF.Close()

	exempF := mustOpen(*exemp)
	defer exempF.Close()

	m := newMain(vmarcF, exempF, outMerged, outNoItems, *limit, *skip)
	if err := m.Run(); err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(filepath.Join(*outDir, "itypes.sql"), []byte(itypesSQL), os.ModePerm); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(*outDir, "avalues.sql"), []byte(aValuesSQL), os.ModePerm); err != nil {
		log.Fatal(err)
	}

	templ := template.Must(template.New("branches").Parse(branchesSQLtmpl))
	branchF := mustCreate(filepath.Join(*outDir, "branches.sql"))
	defer branchF.Close()
	if err := templ.Execute(branchF, m.branches); err != nil {
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
