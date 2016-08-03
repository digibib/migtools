package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
)

const (
	mysqlCountTmpl = `mysql --default-character-set=utf8 -h koha_mysql -u$MYSQL_USER -p$MYSQL_PASSWORD $MYSQL_DATABASE -e %q | tail -n 1`
	sparqlTmpl     = `curl -s -G 'http://fuseki:3030/ds/query?format=csv' --data-urlencode 'query=%s' | tail -n+2`
	esCountTmpl    = `curl -s elasticsearch:9200/search/%s/_count | python -c 'import sys, json; print json.load(sys.stdin)["count"]'`
)

func mysqlCount(q string) string {
	return fmt.Sprintf(mysqlCountTmpl, q)
}

func sparql(q string) string {
	return fmt.Sprintf(sparqlTmpl, q)
}

func esCount(r string) string {
	return fmt.Sprintf(esCountTmpl, r)
}

type ResourceMetric struct {
	Name          string
	Bibliofil     string
	Prepared      string
	Koha          string
	Fuseki        string
	Elasticsearch string
}

type CirculationMetric struct {
	Name      string
	Bibliofil string
	Prepared  string
	Koha      string
}

func init() {
	log.SetFlags(0)
}

var host = "localhost"

func withHost(s string) string {
	return fmt.Sprintf(s, host)
}

func main() {
	hostFlag := flag.String("h", "localhost", "namespace host (for RDF ontology)")
	skipCirc := flag.Bool("nocirc", true, "skip circulation verificaitons")
	flag.Parse()
	host = *hostFlag

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	defer w.Flush()

	resourceChecks := []ResourceMetric{
		{
			Name:          "publications",
			Bibliofil:     `ls -1 /data/*vmarc.*.txt | xargs cat | grep "*001" | wc -l`,
			Prepared:      `cat /out/catalogue.mrc | grep -o $'\035' | wc -l`,
			Koha:          mysqlCount("SELECT COUNT(*) FROM biblioitems"),
			Fuseki:        sparql(withHost("SELECT (COUNT(DISTINCT ?p) AS ?count) WHERE { ?p a <http://%s:8005/ontology#Publication> }")), // TODO use static deichman namespace when ready
			Elasticsearch: esCount("publication"),
		},
		{
			Name:          "items",
			Bibliofil:     `ls -1 /data/*exemp.*.txt | xargs cat | grep "ex_titnr" | wc -l`,
			Prepared:      `cat /out/catalogue.mrc | grep -o 'p0301' | wc -l`,
			Koha:          mysqlCount("SELECT COUNT(*) FROM items"),
			Fuseki:        "",
			Elasticsearch: "",
		},
		{
			Name:          "works",
			Bibliofil:     "",
			Prepared:      `cat /out/resources.nt | grep -o "#Work>" | wc -l`,
			Koha:          "",
			Fuseki:        sparql(withHost("SELECT (COUNT(DISTINCT ?p) AS ?count) WHERE { ?p a <http://%s:8005/ontology#Work> }")), // TODO use static deichman namespace when ready
			Elasticsearch: esCount("work"),
		},
		{
			Name:          "persons",
			Bibliofil:     "",
			Prepared:      `cat /out/resources.nt | grep -o "#Person>" | wc -l`,
			Koha:          "",
			Fuseki:        sparql(withHost("SELECT (COUNT(DISTINCT ?p) AS ?count) WHERE { ?p a <http://%s:8005/ontology#Person> }")), // TODO use static deichman namespace when ready
			Elasticsearch: esCount("person"),
		},
		{
			Name:          "subjects",
			Bibliofil:     "",
			Prepared:      `cat /out/resources.nt | grep -o "#Subject>" | wc -l`,
			Koha:          "",
			Fuseki:        sparql(withHost("SELECT (COUNT(DISTINCT ?p) AS ?count) WHERE { ?p a <http://%s:8005/ontology#Subject> }")), // TODO use static deichman namespace when ready
			Elasticsearch: esCount("subject"),
		},
	}
	fmt.Fprintln(w, "Verifying resources\n===================\n")
	fmt.Fprintln(w, "\tBibliofil\tPrepared\tKoha\tFuseki\tElasticsearch")
	for _, c := range resourceChecks {
		fmt.Fprint(w, c.Name)
		fmt.Fprint(w, "\t")
		for _, source := range []string{c.Bibliofil, c.Prepared, c.Koha, c.Fuseki, c.Elasticsearch} {
			if source == "" {
				fmt.Fprint(w, "-")
			} else {
				cmd := exec.Command("/bin/bash", "-c", source)
				out, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("failed command: %q\n", source)
					log.Println(string(out))
					log.Fatal(err)
				}
				fmt.Fprint(w, strings.TrimSpace(string(out)))
			}
			fmt.Fprint(w, "\t")

		}
		fmt.Fprint(w, "\n")
	}
	fmt.Fprintln(w)

	interestingNumbers := map[string]string{
		"Publications not belonging to any work": withHost(`
		PREFIX : <http://%s:8005/ontology#>
		SELECT COUNT(DISTINCT ?p)
		WHERE {
			?p a :Publication .
			MINUS { ?p :publicationOf ?w .
				    ?w a :Work }
		}`),
		"Works with two MainEntry contributions": withHost(`
		PREFIX : <http://%s:8005/ontology#>
		SELECT COUNT(DISTINCT ?w)
		WHERE {
			?w a :Work ;
			   :contributor ?bnode1 .
			?bnode1 a :MainEntry .
			?w :contributor ?bnode2 .
			?bnode2 a :MainEntry .
			FILTER(?bnode1 != ?bnode2)
		}`),
		"Publications without mediatype": withHost(`
		PREFIX : <http://%s:8005/ontology#>
		SELECT COUNT(DISTINCT ?p)
		WHERE {
			?p a :Publication
			FILTER NOT EXISTS { ?p :mediaType ?mediaType }
		}`),
		"Publications without format": withHost(`
		PREFIX : <http://%s:8005/ontology#>
		SELECT COUNT(DISTINCT ?p)
		WHERE {
			?p a :Publication
			FILTER NOT EXISTS { ?p :format ?format }
		}`),
		"Publication with raw:publicationPlace but not conneted to place of publication": withHost(`
		PREFIX :    <http://%s:8005/ontology#>
		PREFIX raw: <http://data.deichman.no/raw#>
		SELECT COUNT(DISTINCT ?p)
		WHERE {
			?p raw:publicationPlace ?rawPlaceLabel .
			FILTER NOT EXISTS { ?p :hasPlaceOfPublication ?place . }
		}`),
		"Publications without mainTitle": withHost(`
		PREFIX :    <http://%s:8005/ontology#>
		SELECT COUNT(DISTINCT ?p)
		WHERE {
			?p a :Publication .
			FILTER NOT EXISTS { ?p :mainTitle ?mainTitle . }
		}`),
	}

	fmt.Fprintln(w, "\nInteresting numbers\n===================\n")
	for label, q := range interestingNumbers {
		cmd := exec.Command("/bin/sh", "-c", sparql(q))
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("failed command: %q\n", sparql(q))
			log.Println(string(out))
			log.Fatal(err)
		}
		fmt.Fprint(w, strings.TrimSpace(string(out)))
		fmt.Fprint(w, "\t")
		fmt.Fprint(w, label)
		fmt.Fprint(w, "\n")
	}
	fmt.Fprintln(w)

	if *skipCirc {
		return
	}

	circulationChecks := []CirculationMetric{
		{
			Name:      "patrons",
			Bibliofil: "ls -1 /data/*laaner.*.txt | xargs cat | grep ln_nr | wc -l",
			Prepared:  "cat /out/patrons.csv | wc -l",
			Koha:      mysqlCount("SELECT count(*) FROM borrowers"),
		},
		{
			Name:      "issues",
			Bibliofil: `ls -1 /data/*exemp.*.txt | xargs cat | grep "ex_laanr |[^-]" | wc -l`,
			Prepared:  "cat /out/issues.sql | grep INSERT | wc -l",
			Koha:      mysqlCount("SELECT count(*) FROM issues"),
		},
		{
			Name:      "holds",
			Bibliofil: `ls -1 /data/*res.*.txt | xargs cat | grep res_titnr | wc -l`,
			Prepared:  "cat /out/holds.sql | grep INSERT | wc -l",
			Koha:      mysqlCount("SELECT count(*) FROM reserves"),
		},
	}

	fmt.Fprintln(w, "\nVerifying patrons and circulation data\n======================================\n")
	fmt.Fprintln(w, "\tBibliofil\tPrepared\tKoha")
	for _, c := range circulationChecks {
		fmt.Fprint(w, c.Name)
		fmt.Fprint(w, "\t")
		for _, source := range []string{c.Bibliofil, c.Prepared, c.Koha} {
			if source == "" {
				fmt.Fprint(w, "-")
			} else {
				cmd := exec.Command("/bin/sh", "-c", source)
				out, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("failed command: %q\n", source)
					log.Println(string(out))
					log.Fatal(err)
				}
				fmt.Fprint(w, strings.TrimSpace(string(out)))
			}
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, "\n")
	}
	fmt.Fprintln(w)
}
