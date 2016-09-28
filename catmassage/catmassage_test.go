package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/boutros/marc"
)

func parseRecords(t *testing.T, r io.Reader, format marc.Format) []*marc.Record {
	var res []*marc.Record
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
	m := newMain(bytes.NewBufferString(sampleVMARC), bytes.NewReader([]byte(sampleEXEMP)), bytes.NewBufferString(sampleEMARC), &outMerged, &outNoItems, ioutil.Discard, ioutil.Discard, ioutil.Discard, -1, 0)
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
		remove952(want[i])
		if !gotNoItems[i].Eq(want[i]) {
			t.Fatalf("got:\n%+v\nwant:%+v", gotNoItems[i], want[i])
		}
	}

	wantBranchCodes := map[string]string{
		"fbol":   "Bøler",
		"fnyd":   "Nydalen",
		"hutl":   "Hovedbiblioteket, voksen",
		"ffur":   "Furuset",
		"fmaj":   "Majorstuen",
		"ukjent": "Ukjent avdeling",
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
    <datafield tag="019" ind1=" " ind2=" ">
        <subfield code="b">l</subfield>
    </datafield>
    <datafield tag="092" ind1=" " ind2=" ">
        <subfield code="a">MILJØHYLLA</subfield>
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
        <subfield code="y">BOK</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">1</subfield>
        <subfield code="p">03010379371001</subfield>
        <subfield code="a">ukjent</subfield>
        <subfield code="b">ukjent</subfield>
        <subfield code="c">Miljøhylla</subfield>
        <subfield code="l">8</subfield>
        <subfield code="o">641.3 Gra</subfield>
        <subfield code="y">BOK</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">2</subfield>
        <subfield code="p">03010379371002</subfield>
        <subfield code="a">hutl</subfield>
        <subfield code="b">hutl</subfield>
        <subfield code="c">Miljøhylla</subfield>
        <subfield code="1">11</subfield>
        <subfield code="l">14</subfield>
        <subfield code="o">641.3 Gra</subfield>
        <subfield code="y">BOK</subfield>
    </datafield>
</record>
<record>
    <leader>     n   a22        4500</leader>
    <controlfield tag="001">1245593</controlfield>
    <controlfield tag="008">120228                a          00nob 2</controlfield>
    <datafield tag="019" ind1=" " ind2=" ">
        <subfield code="s">18</subfield>
        <subfield code="b">l</subfield>
    </datafield>
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
    <datafield tag="521" ind1=" " ind2=" ">
        <subfield code="a">Aldersgrense 18</subfield>
    </datafield>
    <datafield tag="942" ind1=" " ind2=" ">
        <subfield code="y">BOK</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">1</subfield>
        <subfield code="p">03011245593001</subfield>
        <subfield code="a">hutl</subfield>
        <subfield code="b">hutl</subfield>
        <subfield code="l">23</subfield>
        <subfield code="y">BOK</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">2</subfield>
        <subfield code="p">03011245593002</subfield>
        <subfield code="a">hutl</subfield>
        <subfield code="b">hutl</subfield>
        <subfield code="l">24</subfield>
        <subfield code="y">BOK</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">3</subfield>
        <subfield code="p">03011245593003</subfield>
        <subfield code="a">hutl</subfield>
        <subfield code="b">hutl</subfield>
        <subfield code="l">14</subfield>
        <subfield code="y">DAGSLAAN</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">4</subfield>
        <subfield code="p">03011245593004</subfield>
        <subfield code="a">ffur</subfield>
        <subfield code="b">ffur</subfield>
        <subfield code="c">BEDRE LIV</subfield>
        <subfield code="l">27</subfield>
        <subfield code="y">BOK</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">5</subfield>
        <subfield code="p">03011245593005</subfield>
        <subfield code="a">fmaj</subfield>
        <subfield code="b">fmaj</subfield>
        <subfield code="l">44</subfield>
        <subfield code="y">BOK</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">6</subfield>
        <subfield code="p">03011245593006</subfield>
        <subfield code="a">fbol</subfield>
        <subfield code="b">fbol</subfield>
        <subfield code="c">Livsstil</subfield>
        <subfield code="l">28</subfield>
        <subfield code="y">BOK</subfield>
    </datafield>
    <datafield tag="952" ind1=" " ind2=" ">
        <subfield code="t">7</subfield>
        <subfield code="p">03011245593007</subfield>
        <subfield code="a">fnyd</subfield>
        <subfield code="b">fnyd</subfield>
        <subfield code="l">6</subfield>
        <subfield code="y">BOK</subfield>
    </datafield>
</record>
<record>
    <leader>     c   a22        4500</leader>
    <controlfield tag="001">0192529</controlfield>
    <controlfield tag="008">900326                a          10dan</controlfield>
    <datafield tag="019" ind1=" " ind2=" ">
        <subfield code="b">di|dr</subfield>
    </datafield>
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
        <subfield code="y">LYDBOK</subfield>
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
*019  $bl
*008920916                a          0 nob
*092  $aMILJØHYLLA
*090  $c641.3$dGra
*100 0$aGrahl-Nielsen, Thora$d1901-$jn.$326452400
*24510$aUgress er også mat$bopskrifter og aktuelle surrogater av ville vekster$c[Av] Thora Grahl-Nielsen [og] Astrid Karlsen
^
*000     n
*0011245593
*008120228                a          00nob 2
*019  $s18$bl
*100 0$aMärtha Louise$cprinsesse, datter av Harald V, konge av Norge$d1971-$jn.$1948.1055092$6923.148 z, 948.1055 x$316549700
*24510$aEnglenes hemmeligheter$bderes natur, språk og hvordan du åpner opp for dem$cPrinsesse Märtha Louise, Elisabeth Nordeng
*521  $aNoteThatShouldBeOverwrittenByAgeLimit
^
*000     c
*0010192529
*008900326                a          10dan
*019  $bdi|dr
*100 0$aBrú, Heðin$d1901-1987$jfær.$312827600
*24510$aFjeldskyggen$bnoveller og skitser$cPå dansk ved Gunnvá og Poul Skårup
^
`

const sampleEMARC = `
*0010379371
*0020000001
*015  $aNO:02030000:1003010379371001 ::rfidE004015012C80BFB ::03010379371001
*017  $a36517
*101  $ahutl$d42529
*102  $afgam$d42519
*103  $a30/ 493$l20077216$d42519$k93211$ifgam
*104  $i2016-03-15T15:57:28
*301  $a42443$t155728$va$i2016-03-15T15:57:28
*301  $a42443$t160043$va$i2016-03-15T16:00:43
*301  $a42527$t102457$va$i2016-06-07T10:24:57
^
*0010379371
*0020000003
*017  $a36517
^
*0011245593
*0020000001
*015  $aNO:02030000:1003011245593001 ::rfidE00401500CA3566D ::03011245593001
*017  $a40963
*100  $a42498$t101613$chvua$fLaanerkategori$lNormal
*101  $b2023500$f41927$t15:29:13$ahvua$d42492
*104  $i2016-05-03T14:17:21
*110  $a42526 42561 42582
*301  $a41044$t063230$vV
*301  $a41044$t063232$vV
*301  $a41044$t063238$vV
*301  $a41044$t063247$vV
*301  $a41044$t063249$vV
*301  $a41044$t063249$vV
*301  $a41087$t104918$vV
*301  $a41121$t094203$vV
*301  $a41366$t102011$vV
*301  $a41459$t093835$vV
*301  $a41927$t152913$vI
*301  $a42018$t181850$va
*301  $a42018$t182550$va
^
*0011245593
*0020000002
*015  $aNO:02030000:1003011245593002 ::rfidE00401500CA35794 ::03011245593002
*017  $a40963
*101  $ahvua$d41934
*104  $i2014-10-23T15:04:25
*301  $a41649$t131103$va
*301  $a41677$t144635$va
*301  $a41679$t112925$va
*301  $a41760$t122405$va
*301  $a41760$t122444$va
*301  $a41760$t133558$va
^
*0011245593
*0020000003
*015  $aNO:02030000:1003011245593003 ::rfidE00401500CA35665 ::03011245593003
*017  $a40963
*101  $ahutl$d42428
*104  $i2016-02-26T11:42:07
*250  $aDagslån
*301  $a40988$t101501$vV
*301  $a40988$t101502$vV
*301  $a41085$t120246$vV
*301  $a41085$t120247$vV
*301  $a41085$t120247$vV
*301  $a41085$t120252$vV
*301  $a41085$t120253$vV
*301  $a41085$t120253$vV
*301  $a41704$t093123$va
*301  $a41899$t140334$va
*301  $a41899$t150026$va
*301  $a41900$t064919$vV
*301  $a41900$t064920$vV
*301  $a41900$t064920$vV
*301  $a42425$t114207$va$i2016-02-26T11:42:07
*301  $a42425$t132611$va$i2016-02-26T13:26:11
^
*0011245593
*0020000004
*015  $aNO:02030000:1003011245593004 ::rfidE00401004670A666 ::03011245593004
*016  $a0
*017  $a40963
*101  $affur$d42183
*104  $i2015-05-18T10:27:49
*301  $a41613$t112944$va
*301  $a41856$t104538$va
*301  $a41856$t105555$va
*301  $a42036$t184035$va
*301  $a42037$t090227$va
^
*0011245593
*0020000005
*015  $aNO:02030000:1003011245593005 ::rfidE00401500CA4319A ::03011245593005
*016  $a0
*017  $a40963
*101  $afmaa$d42599
*104  $i2016-08-04T16:39:54
*301  $a41121$t072337$vV
*301  $a41245$t064711$vV
*301  $a41245$t064712$vV
*301  $a41245$t064714$vV
*301  $a41245$t064720$vV
*301  $a41245$t064724$vV
*301  $a41245$t064725$vV
*301  $a41276$t072416$vV
*301  $a41276$t072416$vV
*301  $a41428$t095229$vV
*301  $a41428$t095230$vV
*301  $a41609$t144948$va
*301  $a41609$t152544$va
*301  $a41948$t113900$va
*301  $a41948$t115733$va
*301  $a42040$t132116$va
*301  $a42040$t152451$va
*301  $a42045$t070808$vV
*301  $a42045$t070809$vV
*301  $a42064$t135919$va
*301  $a42064$t142228$va
*301  $a42222$t164618$va$i2015-08-07T16:46:18
*301  $a42222$t170914$va$i2015-08-07T17:09:14
^
*0011245593
*0020000006
*015  $aNO:02030000:1003011245593006 ::rfidE00401004355C55E ::03011245593006
*016  $a0
*017  $a40963
*100  $a42571$t162715$cfboa$fLaanerkategori$lNormal
*101  $afboa$d42165
*104  $i2015-06-11T12:42:25
*110  $a42590
*301  $a41627$t143620$vb
^
*0011245593
*0020000007
*015  $aNO:02030000:1003011245593007 ::rfidE004015022389F9A ::03011245593007
*016  $a0
*017  $a40963
*101  $afnyd$d42424
*102  $afsto$d42418
*104  $i2015-06-10T17:35:08
*301  $a41982$t081027$vV
*301  $a41989$t102324$va
*301  $a41989$t111855$va
*301  $a41990$t061940$vV
*301  $a41990$t061941$vV
*301  $a41990$t061942$vV
*301  $a41990$t061942$vV
*301  $a41990$t061951$vV
*301  $a41990$t061954$vV
*301  $a41990$t061955$vV
*301  $a41990$t061955$vV
*301  $a42109$t121410$va$i2015-04-16T12:14:10
*301  $a42109$t123349$va$i2015-04-16T12:33:49
*301  $a42113$t075123$vV$i2015-04-20T07:51:23
*301  $a42121$t073637$vV$i2015-04-28T07:36:37
*301  $a42164$t173508$va$i2015-06-10T17:35:08
*301  $a42164$t174428$va$i2015-06-10T17:44:28
*301  $a42422$t103350$vV$i2016-02-23T10:33:50
*301  $a42422$t103352$vV$i2016-02-23T10:33:52
*301  $a42422$t111621$va$i2016-02-23T11:16:21
*301  $a42423$t105433$vV$i2016-02-24T10:54:33
*301  $a42423$t105433$vV$i2016-02-24T10:54:33
*301  $a42423$t105434$vV$i2016-02-24T10:54:34
*301  $a42423$t105435$vV$i2016-02-24T10:54:35
*301  $a42424$t074629$vV$i2016-02-25T07:46:29
^
*0010192529
*0020000001
*015  $aNO:02030000:1003010192529001 ::rfidE004015008E19F33 ::03010192529001
*017  $a36517
*101  $b2023000$f42045$t13:48:25$ahutl$d42050
*301  $a42045$t134825$vI
^
*0010192529
*0020000001
*015  $aNO:02030000:1003010192529001 ::rfidE004015008E19F33 ::03010192529001
*017  $a36517
*101  $b2023000$f42045$t13:48:25$ahutl$d42050
*301  $a42045$t134825$vI
^
`

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
	fbjl, fnyl := splitItems(r)

	if len(fbjl) != 1 || len(fnyl) != 2 {
		t.Errorf("bjønrholt/nydalen items, got %d/%d; want 1/2", len(fbjl), len(fnyl))
	}
}
