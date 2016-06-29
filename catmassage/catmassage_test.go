package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/boutros/marc"
)

func parseRecords(t *testing.T, r io.Reader, format marc.Format) []marc.Record {
	var res []marc.Record
	dec := marc.NewDecoder(r, format)
	for r, err := dec.Decode(); err != io.EOF; r, err = dec.Decode() {
		if err != nil {
			t.Fatal(err)
		}
		res = append(res, r)
	}
	return res
}

func TestMerge(t *testing.T) {
	var outMerged bytes.Buffer
	var outNoItems bytes.Buffer
	m := newMain(bytes.NewBufferString(sampleVMARC), bytes.NewReader([]byte(sampleEXEMP)), &outMerged, &outNoItems, ioutil.Discard, ioutil.Discard, -1, 0)
	if err := m.Run(); err != nil {
		t.Fatal(err)
	}

	got := parseRecords(t, bytes.NewReader(outMerged.Bytes()), marc.MARC)
	gotNoItems := parseRecords(t, bytes.NewReader(outNoItems.Bytes()), marc.MARCXML)
	want := parseRecords(t, bytes.NewBufferString(wantMARCXML), marc.MARCXML)

	if len(got) != len(want) {
		t.Fatalf("got:\n%v\nwant:%v", got, want)
	}
	if len(gotNoItems) != len(got) {
		t.Fatalf("number of marcxml records %d != %d number of marc records", len(gotNoItems), len(got))
	}
	for i, r := range got {
		if !r.Eq(want[i]) {
			t.Fatalf("got:\n%+v\nwant:%+v", r, want[i])
		}
		// verify that the full marcxml and marc records without items are equal, when the 952 fields are removed:
		remove952(&want[i])
		if !gotNoItems[i].Eq(want[i]) {
			t.Fatalf("got:\n%+v\nwant:%+v", gotNoItems[i], want[i])
		}
	}

	wantBranchCodes := map[string]string{
		"fbol": "Bøler",
		"fnyd": "Nydalen",
		"hutl": "Hovedbiblioteket",
		"ffur": "Furuset",
		"fmaj": "Majorstua",
		"xyz":  "MISSING LABEL FOR BRANCH \"xyz\"",
	}
	if !reflect.DeepEqual(wantBranchCodes, m.branches) {
		t.Fatalf("got:\n%v\nwant:\n%v", wantBranchCodes, m.branches)
	}
}

const wantMARCXML = `<?xml version="1.0" encoding="UTF-8"?>
<collection xmlns="http://www.loc.gov/MARC21/slim">
<record>
    <leader>     c   a22        4500</leader>
    <controlfield tag="001">0379371</controlfield>
    <controlfield tag="008">920916                a          0 nob</controlfield>
    <datafield tag="090" ind1=" " ind2=" ">
        <subfield code="c">641.3</subfield>
        <subfield code="d">Gra</subfield>
    </datafield>
    <datafield tag="100" ind1=" " ind2="0">
        <subfield code="a">Grahl-Nielsen, Thora</subfield>
        <subfield code="d">1901-</subfield>
        <subfield code="j">n.</subfield>
        <subfield code="3">26452400</subfield>
    </datafield>
    <datafield tag="245" ind1="1" ind2="0">
        <subfield code="a">Ugress er også mat</subfield>
        <subfield code="b">opskrifter og aktuelle surrogater av ville vekster</subfield>
        <subfield code="c">[Av] Thora Grahl-Nielsen [og] Astrid Karlsen</subfield>
    </datafield>
    <datafield tag="942" ind1=" " ind2=" ">
        <subfield code="y">X</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">1</subfield>
        <subfield code="p">03010379371001</subfield>
        <subfield code="a">xyz</subfield>
        <subfield code="b">xyz</subfield>
        <subfield code="c">m</subfield>
        <subfield code="l">8</subfield>
        <subfield code="o">641.3 Gra</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">2</subfield>
        <subfield code="p">03010379371002</subfield>
        <subfield code="a">hutl</subfield>
        <subfield code="b">hutl</subfield>
        <subfield code="c">m</subfield>
        <subfield code="1">11</subfield>
        <subfield code="l">14</subfield>
        <subfield code="o">641.3 Gra</subfield>
        <subfield code="y">28</subfield>
    </datafield>
</record>
<record>
    <leader>     n   a22        4500</leader>
    <controlfield tag="001">1245593</controlfield>
    <controlfield tag="008">120228                a          00nob 2</controlfield>
    <datafield tag="100" ind1=" " ind2="0">
        <subfield code="a">Märtha Louise</subfield>
        <subfield code="c">prinsesse, datter av Harald V, konge av Norge</subfield>
        <subfield code="d">1971-</subfield>
        <subfield code="j">n.</subfield>
        <subfield code="1">948.1055092</subfield>
        <subfield code="6">923.148 z, 948.1055 x</subfield>
        <subfield code="3">16549700</subfield>
    </datafield>
    <datafield tag="245" ind1="1" ind2="0">
        <subfield code="a">Englenes hemmeligheter</subfield>
        <subfield code="b">deres natur, språk og hvordan du åpner opp for dem</subfield>
        <subfield code="c">Prinsesse Märtha Louise, Elisabeth Nordeng</subfield>
    </datafield>
    <datafield tag="942" ind1=" " ind2=" ">
        <subfield code="y">X</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">1</subfield>
        <subfield code="p">03011245593001</subfield>
        <subfield code="a">hutl</subfield>
        <subfield code="b">hutl</subfield>
        <subfield code="l">23</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">2</subfield>
        <subfield code="p">03011245593002</subfield>
        <subfield code="a">hutl</subfield>
        <subfield code="b">hutl</subfield>
        <subfield code="l">24</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">3</subfield>
        <subfield code="p">03011245593003</subfield>
        <subfield code="a">hutl</subfield>
        <subfield code="b">hutl</subfield>
        <subfield code="l">14</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">4</subfield>
        <subfield code="p">03011245593004</subfield>
        <subfield code="a">ffur</subfield>
        <subfield code="b">ffur</subfield>
        <subfield code="c">BEDRE LIV</subfield>
        <subfield code="l">27</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">5</subfield>
        <subfield code="p">03011245593005</subfield>
        <subfield code="a">fmaj</subfield>
        <subfield code="b">fmaj</subfield>
        <subfield code="l">44</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">6</subfield>
        <subfield code="p">03011245593006</subfield>
        <subfield code="a">fbol</subfield>
        <subfield code="b">fbol</subfield>
        <subfield code="c">Livsstil</subfield>
        <subfield code="l">28</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">7</subfield>
        <subfield code="p">03011245593007</subfield>
        <subfield code="a">fnyd</subfield>
        <subfield code="b">fnyd</subfield>
        <subfield code="l">6</subfield>
        <subfield code="y">28</subfield>
    </datafield>
</record>
<record>
    <leader>     c   a22        4500</leader>
    <controlfield tag="001">0192529</controlfield>
    <controlfield tag="008">900326                a          10dan</controlfield>
    <datafield tag="100" ind1=" " ind2="0">
        <subfield code="a">Brú, Heðin</subfield>
        <subfield code="d">1901-1987</subfield>
        <subfield code="j">fær.</subfield>
        <subfield code="3">12827600</subfield>
    </datafield>
    <datafield tag="245" ind1="1" ind2="0">
        <subfield code="a">Fjeldskyggen</subfield>
        <subfield code="b">noveller og skitser</subfield>
        <subfield code="c">På dansk ved Gunnvá og Poul Skårup</subfield>
    </datafield>
    <datafield tag="942" ind1=" " ind2=" ">
        <subfield code="y">X</subfield>
    </datafield>
</record>
</collection>`

const sampleVMARC = `
*000     d
*0010379371
*24510$aXXX
^
*000     c
*0010379371
*008920916                a          0 nob
*090  $c641.3$dGra
*100 0$aGrahl-Nielsen, Thora$d1901-$jn.$326452400
*24510$aUgress er også mat$bopskrifter og aktuelle surrogater av ville vekster$c[Av] Thora Grahl-Nielsen [og] Astrid Karlsen
^
*000     n
*0011245593
*008120228                a          00nob 2
*100 0$aMärtha Louise$cprinsesse, datter av Harald V, konge av Norge$d1971-$jn.$1948.1055092$6923.148 z, 948.1055 x$316549700
*24510$aEnglenes hemmeligheter$bderes natur, språk og hvordan du åpner opp for dem$cPrinsesse Märtha Louise, Elisabeth Nordeng
^
*000     c
*0010192529
*008900326                a          10dan
*100 0$aBrú, Heðin$d1901-1987$jfær.$312827600
*24510$aFjeldskyggen$bnoveller og skitser$cPå dansk ved Gunnvá og Poul Skårup
^`

const sampleEXEMP = `
ex_titnr |379371|
ex_exnr |1|
ex_avd |xyz|
ex_plass |m|
ex_hylle ||
ex_note ||
ex_bind |0|
ex_aar |0|
ex_status ||
ex_resstat ||
ex_laanstat ||
ex_utlkode ||
ex_laanr |-3034969|
ex_laantid |28|
ex_forfall |00/00/0000|
ex_purrdat |00/00/0000|
ex_antpurr |0|
ex_etikett ||
ex_antlaan |8|
ex_kl_sett |0|
ex_strek |0|
^
ex_titnr |379371|
ex_exnr |2|
ex_avd |hutl|
ex_plass |m|
ex_hylle ||
ex_note ||
ex_bind |0|
ex_aar |0|
ex_status |y|
ex_resstat ||
ex_laanstat ||
ex_utlkode ||
ex_laanr |2|
ex_laantid |28|
ex_forfall |00/00/0000|
ex_purrdat |00/00/0000|
ex_antpurr |0|
ex_etikett ||
ex_antlaan |14|
ex_kl_sett |0|
ex_strek |0|
^
ex_titnr |1245593|
ex_exnr |1|
ex_avd |hutl|
ex_plass ||
ex_hylle ||
ex_note ||
ex_bind |0|
ex_aar |2012|
ex_status ||
ex_resstat ||
ex_laanstat ||
ex_utlkode ||
ex_laanr |-9964745|
ex_laantid |28|
ex_forfall |00/00/0000|
ex_purrdat |00/00/0000|
ex_antpurr |0|
ex_etikett ||
ex_antlaan |23|
ex_kl_sett |0|
ex_strek |0|
^
ex_titnr |1245593|
ex_exnr |2|
ex_avd |hutl|
ex_plass ||
ex_hylle ||
ex_note ||
ex_bind |0|
ex_aar |2012|
ex_status ||
ex_resstat ||
ex_laanstat ||
ex_utlkode ||
ex_laanr |-9964746|
ex_laantid |28|
ex_forfall |00/00/0000|
ex_purrdat |00/00/0000|
ex_antpurr |0|
ex_etikett ||
ex_antlaan |24|
ex_kl_sett |0|
ex_strek |0|
^
ex_titnr |1245593|
ex_exnr |3|
ex_avd |hutl|
ex_plass ||
ex_hylle ||
ex_note ||
ex_bind |0|
ex_aar |2012|
ex_status ||
ex_resstat ||
ex_laanstat ||
ex_utlkode ||
ex_laanr |-9964747|
ex_laantid |28|
ex_forfall |00/00/0000|
ex_purrdat |00/00/0000|
ex_antpurr |0|
ex_etikett ||
ex_antlaan |14|
ex_kl_sett |0|
ex_strek |0|
^
ex_titnr |1245593|
ex_exnr |4|
ex_avd |ffur|
ex_plass |BEDRE LIV|
ex_hylle ||
ex_note ||
ex_bind |0|
ex_aar |2012|
ex_status ||
ex_resstat ||
ex_laanstat ||
ex_utlkode ||
ex_laanr |-9964748|
ex_laantid |28|
ex_forfall |00/00/0000|
ex_purrdat |00/00/0000|
ex_antpurr |0|
ex_etikett ||
ex_antlaan |27|
ex_kl_sett |0|
ex_strek |-1245593|
^
ex_titnr |1245593|
ex_exnr |5|
ex_avd |fmaj|
ex_plass ||
ex_hylle ||
ex_note ||
ex_bind |0|
ex_aar |2012|
ex_status ||
ex_resstat ||
ex_laanstat ||
ex_utlkode ||
ex_laanr |-9964749|
ex_laantid |28|
ex_forfall |00/00/0000|
ex_purrdat |00/00/0000|
ex_antpurr |0|
ex_etikett ||
ex_antlaan |44|
ex_kl_sett |0|
ex_strek |-1245593|
^
ex_titnr |1245593|
ex_exnr |6|
ex_avd |fbol|
ex_plass |Livsstil|
ex_hylle ||
ex_note ||
ex_bind |0|
ex_aar |2012|
ex_status ||
ex_resstat ||
ex_laanstat ||
ex_utlkode ||
ex_laanr |-9964750|
ex_laantid |28|
ex_forfall |00/00/0000|
ex_purrdat |00/00/0000|
ex_antpurr |0|
ex_etikett ||
ex_antlaan |28|
ex_kl_sett |0|
ex_strek |-1245593|
^
ex_titnr |1245593|
ex_exnr |7|
ex_avd |fnyd|
ex_plass ||
ex_hylle ||
ex_note ||
ex_bind |0|
ex_aar |0|
ex_status ||
ex_resstat ||
ex_laanstat ||
ex_utlkode ||
ex_laanr |-9964751|
ex_laantid |28|
ex_forfall |00/00/0000|
ex_purrdat |00/00/0000|
ex_antpurr |0|
ex_etikett ||
ex_antlaan |6|
ex_kl_sett |0|
ex_strek |-1245593|
^
ex_titnr |192529|
ex_exnr |1|
ex_avd |dfb|
ex_plass |m|
ex_hylle ||
ex_note ||
ex_bind |0|
ex_aar |0|
ex_status ||
ex_resstat ||
ex_laanstat ||
ex_utlkode ||
ex_laanr |-1540233|
ex_laantid |28|
ex_forfall |00/00/0000|
ex_purrdat |00/00/0000|
ex_antpurr |0|
ex_etikett ||
ex_antlaan |8|
ex_kl_sett |0|
ex_strek |0|
^
`

const recordToSplit = `
<record>
    <leader>     n   a22        4500</leader>
    <controlfield tag="001">1245593</controlfield>
    <controlfield tag="008">120228                a          00nob 2</controlfield>
    <datafield tag="100" ind1=" " ind2="0">
        <subfield code="a">Märtha Louise</subfield>
        <subfield code="c">prinsesse, datter av Harald V, konge av Norge</subfield>
        <subfield code="d">1971-</subfield>
        <subfield code="j">n.</subfield>
        <subfield code="1">948.1055092</subfield>
        <subfield code="6">923.148 z, 948.1055 x</subfield>
        <subfield code="3">16549700</subfield>
    </datafield>
    <datafield tag="245" ind1="1" ind2="0">
        <subfield code="a">Englenes hemmeligheter</subfield>
        <subfield code="b">deres natur, språk og hvordan du åpner opp for dem</subfield>
        <subfield code="c">Prinsesse Märtha Louise, Elisabeth Nordeng</subfield>
    </datafield>
    <datafield tag="942" ind1=" " ind2=" ">
        <subfield code="y">X</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">1</subfield>
        <subfield code="p">03011245593001</subfield>
        <subfield code="a">hutl</subfield>
        <subfield code="b">hutl</subfield>
        <subfield code="l">23</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">2</subfield>
        <subfield code="p">03011245593002</subfield>
        <subfield code="a">fbjl</subfield>
        <subfield code="b">fbjl</subfield>
        <subfield code="l">24</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">3</subfield>
        <subfield code="p">03011245593003</subfield>
        <subfield code="a">hutl</subfield>
        <subfield code="b">hutl</subfield>
        <subfield code="l">14</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">4</subfield>
        <subfield code="p">03011245593004</subfield>
        <subfield code="a">ffur</subfield>
        <subfield code="b">ffur</subfield>
        <subfield code="c">BEDRE LIV</subfield>
        <subfield code="l">27</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">5</subfield>
        <subfield code="p">03011245593005</subfield>
        <subfield code="a">fnyl</subfield>
        <subfield code="b">fnyl</subfield>
        <subfield code="l">44</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">6</subfield>
        <subfield code="p">03011245593006</subfield>
        <subfield code="a">fbol</subfield>
        <subfield code="b">fbol</subfield>
        <subfield code="c">Livsstil</subfield>
        <subfield code="l">28</subfield>
        <subfield code="y">28</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">7</subfield>
        <subfield code="p">03011245593007</subfield>
        <subfield code="a">fnyl</subfield>
        <subfield code="b">fnyl</subfield>
        <subfield code="l">6</subfield>
        <subfield code="y">28</subfield>
    </datafield>
</record>`

func TestSplitItems(t *testing.T) {
	r := parseRecords(t, bytes.NewBufferString(recordToSplit), marc.MARCXML)[0]
	fbjl, fnyl := splitItems(&r)

	if len(fbjl) != 1 || len(fnyl) != 2 {
		t.Errorf("bjønrholt/nydalen items, got %d/%d; want 1/2", len(fbjl), len(fnyl))
	}
}
