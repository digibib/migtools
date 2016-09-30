package main

import (
	"bytes"
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

// SPARQL queries
const queries = `
# tag: selectResourceURIs
WITH <http://deichman.no/migration>
SELECT DISTINCT ?uri WHERE {
	?uri a <{{.Type}}> .
}

# tag: constructResource
WITH <http://deichman.no/migration>
CONSTRUCT WHERE {
	<{{.URI}}> ?p ?o
}

# tag: constructWork
PREFIX : <http://data.deichman.no/ontology#>
WITH <http://deichman.no/migration>
CONSTRUCT {
	<{{.URI}}> a :Work ;
			?pred ?obj .
}
WHERE {
	<{{.URI}}> a :Work ;
			?pred ?obj .
	VALUES ?pred { :mainTitle
				   :partTitle
				   :literaryForm
				   :audience
				   :language
				   :genre
				   :subject
				   :mediaType
				   :publicationYear
				   :hasCompositionType
				 }
}

# tag: constructWorkContributions
PREFIX     : <http://data.deichman.no/ontology#>
PREFIX role: <http://data.deichman.no/role#>
WITH <http://deichman.no/migration>
CONSTRUCT {
	<{{.URI}}> :contributor [
		:agent ?agent ;
		:role ?role ;
		a :Contribution ] .
}
WHERE {
	SELECT DISTINCT ?agent ?role WHERE {
		?pub :publicationOf <{{.URI}}> .
		?pub ?role ?agent .
		VALUES ?role {
			role:scriptWriter
			role:actor
			role:composer
			role:director
			role:author
			role:editor
			role:lyricist
			role:adaptor
		}
		FILTER NOT EXISTS { ?pub :mainEntry ?agent }
	}
}

# tag: constructWorkMainEntryContribution
PREFIX     : <http://data.deichman.no/ontology#>
PREFIX role: <http://data.deichman.no/role#>
PREFIX migration: <http://migration.deichman.no/>
WITH <http://deichman.no/migration>
CONSTRUCT {
	<{{.URI}}> :contributor [
		:agent ?agent ;
		:role ?role ;
		a :Contribution, :MainEntry ] .
}
WHERE {
	SELECT DISTINCT ?agent ?role WHERE {
		?pub :publicationOf <{{.URI}}> ;
		     :mainEntry ?agent ;
			 migration:mainEntryRole ?role .
	} LIMIT 1
}

# tag: constructWorkClassifications
PREFIX          : <http://data.deichman.no/ontology#>
PREFIX migration: <http://migration.deichman.no/>
PREFIX       raw: <http://data.deichman.no/raw#>
WITH <http://deichman.no/migration>
CONSTRUCT {
	<{{.URI}}> :hasClassification [
		:hasClassificationNumber ?dewey ;
		:hasClassificationSource ?deweyEdition ;
		a :ClassificationEntry ] .
}
WHERE {
	SELECT DISTINCT ?dewey ?deweyEdition WHERE {
		?pub :publicationOf <{{.URI}}> ;
			migration:classification ?class .
		?class raw:classificationNotation ?dewey .
		OPTIONAL { ?class raw:classificationEdition ?classEdition .
			       VALUES (?classEdition ?deweyEdition) {
			              ("5"      <http://data.deichman.no/classificationSource#ddk5>)
			              ("d5"     <http://data.deichman.no/classificationSource#ddk5>)
			              ("d23"    <http://data.deichman.no/classificationSource#ddk23>)
			              ("23/nor" <http://data.deichman.no/classificationSource#ddk23>)
			              ("DDC-23" <http://data.deichman.no/classificationSource#ddc23>)
			              ("DDK-5 " <http://data.deichman.no/classificationSource#ddk5>)
			              ("22"     <http://data.deichman.no/classificationSource#ddc22>)
			              ("DDK-5"  <http://data.deichman.no/classificationSource#ddk5>)
			              ("DDC-22" <http://data.deichman.no/classificationSource#ddc22>)
			              ("23/no"  <http://data.deichman.no/classificationSource#ddk23>)
			              ("23nor"  <http://data.deichman.no/classificationSource#ddk23>)
			              ("23/mus" <http://data.deichman.no/classificationSource#ddk23>)
			              ("5.utg." <http://data.deichman.no/classificationSource#ddk5>)
			              ("23/n"   <http://data.deichman.no/classificationSource#ddk23>)
			        }
			    }
	}
}

# tag: constructPublication
PREFIX          : <http://data.deichman.no/ontology#>
PREFIX      role: <http://data.deichman.no/role#>
PREFIX migration: <http://migration.deichman.no/>
WITH <http://deichman.no/migration>
CONSTRUCT {
	<{{.URI}}> ?p ?o ; a :Publication .
}
WHERE {
	<{{.URI}}> ?p ?o .
	MINUS {
		<{{.URI}}> ?p ?o .
		VALUES ?p {
			:mainEntry
			migration:series
			migration:seriesEntry
			role:scriptWriter
			role:actor
			role:photographer
			role:lyricist
			role:composer
			role:director
			role:performer
			role:musicalArranger
			role:reader
			role:conductor
			role:author
			role:translator
			role:illustrator
			role:editor
			role:contributor
			role:coreographer
			role:publisher
			role:adaptor
		}
	}
}

# tag: constructPublicationContributions
PREFIX     : <http://data.deichman.no/ontology#>
PREFIX role: <http://data.deichman.no/role#>
WITH <http://deichman.no/migration>
CONSTRUCT {
	<{{.URI}}> :contributor [
		:agent ?agent ;
		:role ?role ;
		a :Contribution ] .
}
WHERE {
	SELECT DISTINCT ?agent ?role WHERE {
	<{{.URI}}> a :Publication ;
			   ?role ?agent .
	VALUES ?role {
		role:photographer
		role:performer
		role:musicalArranger
		role:reader
		role:conductor
		role:translator
		role:illustrator
		role:contributor
		role:coreographer
		role:publisher
		}
	} LIMIT 1
}

# tag: constructPublicationSerials
PREFIX          : <http://data.deichman.no/ontology#>
PREFIX       raw: <http://data.deichman.no/raw#>
PREFIX migration: <http://migration.deichman.no/>
WITH <http://deichman.no/migration>
CONSTRUCT {
	<{{.URI}}> :inSerial [
		a :SerialIssue ;
		:serial ?serial ;
		:issue ?numInSerial ] .
}
WHERE {
	SELECT DISTINCT ?serial ?numInSerial WHERE {
		<{{.URI}}> migration:seriesEntry ?serialEntry .
		?serialEntry migration:series ?serial .
		OPTIONAL { ?serialEntry raw:volumeNumber ?numInSerial . }
	}
}

`

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
			},
			"work": {
				"constructWork",
				"constructWorkMainEntryContribution",
				"constructWorkContributions",
				"constructWorkClassifications",
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

func init() {
	queryBank = sparql.LoadBank(bytes.NewBufferString(queries))
}

func main() {
	services := flag.String("se", "http://localhost:8005", "services endpoint")
	virtuoso := flag.String("ve", "http://localhost:8890/sparql", "virtuoso endpoint")
	numWorkers := flag.Int("n", 3, "number of workers")
	limit := flag.Int("l", -1, "limit to n resources")

	flag.Parse()

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
