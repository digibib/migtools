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

var resourceChecks = []ResourceMetric{
	{
		Name:          "publications",
		Bibliofil:     `ls -1 /data/*vmarc.*.txt | xargs cat | grep "*001" | wc -l`,
		Prepared:      `cat /out/catalogue.mrc | grep -o $'\035' | wc -l`,
		Koha:          mysqlCount("SELECT COUNT(*) FROM biblioitems"),
		Fuseki:        sparql(withHost("SELECT (COUNT(DISTINCT ?p) AS ?count) WHERE { ?p a <http://%s:8005/ontology#Publication> }")), // TODO use static deichman namespace when ready
		Elasticsearch: esCount("publication"),
	},
	{
		Name:          "works",
		Bibliofil:     "",
		Prepared:      `cat /out/resources.nt | grep -o "#Work>" | wc -l`,
		Koha:          "",
		Fuseki:        sparql(withHost("SELECT (COUNT(DISTINCT ?p) AS ?count) WHERE { ?p a <http://%s:8005/ontology#Work> }")), // TODO use static deichman namespace when ready
		Elasticsearch: esCount("work"),
	},
}

var circulationChecks = []CirculationMetric{
	{
		Name:      "patrons",
		Bibliofil: "ls -1 /data/*laaner.*.txt | xargs cat | grep ln_nr | wc -l",
		Prepared:  "cat /out/patrons.csv | wc -l",
		Koha:      mysqlCount("SELECT count(*) FROM borrowers"),
	},
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
	flag.Parse()
	host = *hostFlag

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)
	defer w.Flush()

	fmt.Fprintln(w, "Verifying resources")
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

	fmt.Fprintln(w, "\nVerifying circulation data")
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
	}
	fmt.Fprintln(w)
}
