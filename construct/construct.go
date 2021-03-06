package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/knakk/rdf"
	"github.com/knakk/sparql"
)

var queryBank sparql.Bank

// Main represents the main program execution.
type Main struct {
	services string
	virtuoso *sparql.Repo

	enc      *rdf.TripleEncoder
	wg       sync.WaitGroup      // keep track of completed jobs
	jobs     chan (Resource)     // channel of resources to be processed
	complete chan ([]rdf.Triple) // channel of complete resources to be written out
	bnodeID  uint64              // blank node ID counter
	limit    int
}

// Resource represent a job of migrating a RDF resource
type Resource struct {
	URI  rdf.IRI
	Type string       // person, work or publication
	Old  []rdf.Triple // triples returned from virtuoso
	New  []rdf.Triple // triples to be migrated

}

func newMain(se, ve string, limit int) *Main {
	repo, err := sparql.NewRepo(ve)
	if err != nil {
		log.Fatal(err)
	}
	return &Main{
		services: se,
		virtuoso: repo,
		jobs:     make(chan Resource),
		complete: make(chan []rdf.Triple),
		limit:    limit,
	}
}

func mustBlank(s string) rdf.Blank {
	b, err := rdf.NewBlank(s)
	if err != nil {
		panic(err)
	}
	return b
}

func (m *Main) ensureUniqueBNodeIDs(tr []rdf.Triple) {
	bnodes := make(map[rdf.Blank]rdf.Blank)
	for i, t := range tr {
		if t.Subj.Type() == rdf.TermBlank {
			id := t.Subj.(rdf.Blank)
			if _, ok := bnodes[id]; !ok {
				newId := atomic.AddUint64(&m.bnodeID, 1)
				bnodes[id] = mustBlank(strconv.FormatUint(newId, 10))
			}
			tr[i].Subj = bnodes[id]
		}
		if t.Obj.Type() == rdf.TermBlank {
			id := t.Obj.(rdf.Blank)
			if _, ok := bnodes[id]; !ok {
				newId := atomic.AddUint64(&m.bnodeID, 1)
				bnodes[id] = mustBlank(strconv.FormatUint(newId, 10))
			}
			tr[i].Obj = bnodes[id]
		}
	}
}

// process jobs from jobs channel
func (m *Main) processResources() {

getJob:
	for job := range m.jobs {
		typeQueries := map[string][]string{
			"publication": {
				"constructPublication",
				"constructPublicationContributions",
				"constructPublicationSerials",
				"constructPublicationHasPublicationPart",
				"constructPublicationClassifications",
			},
			"work": {
				"constructResource",
				"constructWorkMainEntryContribution",
				"constructWorkContributions",
				"constructWorkBasedOn",
				"constructWorkFollows",
				"constructWorkContinues",
				"constructWorkIsPartOfWork",
			},
			"person": {
				"constructResource",
			},
			"serial": {
				"constructResource",
			},
			"genre": {
				"constructResource",
			},
			"subject": {
				"constructResource",
			},
			"place": {
				"constructResource",
			},
			"corporation": {
				"constructResource",
			},
			"compositionType": {
				"constructResource",
			},
		}

		var triples []rdf.Triple
		for _, query := range typeQueries[job.Type] {
			q, err := queryBank.Prepare(
				query, struct{ URI string }{job.URI.String()})
			if err != nil {
				log.Fatal(err)
			}
			tr, err := m.virtuoso.Construct(q)
			if err != nil {
				log.Println(err)
				log.Printf("putting job back on queue: %v", job.URI)
				go func() {
					m.jobs <- job
				}()
				continue getJob
			}
			m.ensureUniqueBNodeIDs(tr)
			triples = append(triples, tr...)
		}

		job.Old = triples
		job.New = make([]rdf.Triple, len(job.Old))
		copy(job.New, job.Old)

		stripGyearTimeZone(job.New)

		m.wg.Add(1)
		m.complete <- job.New
	}
	m.wg.Done()
}

func stripGyearTimeZone(ts []rdf.Triple) {
	g, _ := rdf.NewIRI("http://www.w3.org/2001/XMLSchema#gYear")
	for i, t := range ts {
		if term, ok := t.Obj.(rdf.Literal); ok {
			if rdf.TermsEqual(term.DataType, g) {
				s := term.String()
				y := strings.TrimSuffix(s, "Z")
				lit := rdf.NewTypedLiteral(y, g)
				ts[i].Obj = lit
			}

		}
	}
}

func (m *Main) addToQueue(resource string) {
	resourceType := fmt.Sprintf("http://data.deichman.no/ontology#%s", strings.Title(resource))

	q, err := queryBank.Prepare(
		"selectResourceURIs",
		struct{ Type string }{resourceType},
	)
	if err != nil {
		log.Fatal(err)
	}
	res, err := m.virtuoso.Query(q)
	if err != nil {
		log.Fatal(err)
	}

	// loop uri response and add to job queue
	c := 0
	for _, b := range res.Results.Bindings {
		uri, _ := rdf.NewIRI(b["uri"].Value)
		r := Resource{URI: uri, Type: resource}
		m.jobs <- r
		c++
		if c == m.limit {
			break
		}
	}
}

// Writer serializes the completed resources to stdout
func (m *Main) Writer() {
	for tr := range m.complete {
		if err := m.enc.EncodeAll(tr); err != nil {
			log.Fatal(err)
		}
		m.wg.Done()
	}
}

// Run executes the migration process
func (m *Main) Run(workers int) {
	m.enc = rdf.NewTripleEncoder(os.Stdout, rdf.NTriples)
	defer m.enc.Close()
	go m.Writer()
	m.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go m.processResources()
	}
	for _, r := range []string{"person", "corporation", "work", "serial", "genre", "subject", "place", "publication", "compositionType"} {
		m.addToQueue(r)
	}
	close(m.jobs)
	m.wg.Wait()
	close(m.complete)
}

func main() {
	queryFile := flag.String("qf", "/out/constructs.sparql", "sparql queries fiel")
	services := flag.String("se", "http://localhost:8005", "services endpoint")
	virtuoso := flag.String("ve", "http://localhost:8890/sparql", "virtuoso endpoint")
	numWorkers := flag.Int("n", 3, "number of workers")
	limit := flag.Int("l", -1, "limit to n resources")

	flag.Parse()

	q, err := os.Open(*queryFile)
	if err != nil {
		log.Fatal(err)
	}
	queryBank = sparql.LoadBank(q)
	q.Close()

	vURL, err := url.Parse(*virtuoso)
	if err != nil {
		log.Fatal(err)
	}
	host, port, _ := net.SplitHostPort(vURL.Host)
	virtuosoIP, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		log.Fatal(err)
	}

	m := newMain(*services, fmt.Sprintf("http://%s:%s/%s", virtuosoIP.String(), port, vURL.Path), *limit)

	m.Run(*numWorkers)
}
