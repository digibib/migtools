package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/boutros/marc"
)

const (
	laanerDump = `ln_nr |808708|
ln_navn |Testesen, Test|
ln_adr1 |Testgata 12|
ln_adr2 ||
ln_post |0475 OSLO|
ln_land |no|
ln_sprog ||
ln_tlf ||
ln_kat |v|
ln_arbg |Furukneika 2|
ln_altadr |38086944|
ln_altpost |4622 KRISTIANSAND S|
ln_foedt |02/03/1911|
ln_kjoenn |m|
ln_friobs ||
ln_obs ||
ln_regnsendt |00/00/0000|
ln_melding ||
ln_kortdato |11/01/2002|
ln_sistelaan |08/06/2016|
ln_sperres |00/00/0000|
ln_antlaan |117|
ln_antpurr |0|
ln_alt_id ||
^
`
	lmarcDump = `*0010808708
*102  $a2015-06-30T13:24:44$kMappaMi
*103  $a2015-06-30T13:24:46$kMappaMi
*140  $ahutl$bhutl
*200  $s0
*240  $a99887766$cmobilsms
*250  $s0
*254  $a2030000
*260  $l26/10/2010 16:28:57$c196
*261  $z9a925d1cebb962b1629f75f2540bbde0$a1234$b01/07/2015$i2015-07-01T01:11:43$p6803$kflnrlib.tcl:flnr_felles2lokal
*270  $aepost
*271  $aepost
*272  $aEPost
*273  $aUbestemt$bUbestemt
*300  $a1
*400  $dno$s0$e0/0/0
*402  $aWaldemar Thranes gate 42 B$c0171 OSLO$lno$d2015-06-30T13:24:44
*517  $a0
*518  $a0
*526  $afnrmld1$b2
*600  $aN001600007$bN001500002$k1
*601  $aTestesen, Test$bTestgata 12$d0475$fno$g0
*602  $eno$f0$g0000-00-00
*603  $c41630676
*604  $atesttestesen@gmail.com$b0
*605  $b2030000
*606  $a1981-08-09$b02031145555$cM$z9a925d1cebb962b1629f75f2540bbde0$d1234$f0
*607  $a2008-07-18T11:21:25$b2100100$c2015-06-30T13:24:45$d2030000
^
`

	lnelDump = `lnel_nr |808708|
lnel_epost |testtestesen@gmail.com|
lnel_pin |9999|
^
`
)

func TestPatronMerge(t *testing.T) {
	want := patron{
		cardnumber:             "N001600007",
		surname:                "Testesen",
		firstname:              "Test",
		address:                "Testgata 12",
		city:                   "OSLO",
		zipcode:                "0475",
		country:                "no",
		email:                  "testtestesen@gmail.com",
		phone:                  "",
		smsalertnumber:         "99887766",
		userid:                 "808708",
		dateenrolled:           "2002-01-11",
		dateofbirth:            "1911-03-02",
		dateexpiry:             "2099-01-01",
		branchcode:             "hutl",
		categorycode:           "v", // mapped to "V" in Main.Run()
		TEMP_personnr:          "02031145555",
		TEMP_nl:                true,
		TEMP_res_transport:     "epost",
		TEMP_pur_transport:     "epost",
		TEMP_fvarsel_transport: "epost",
		TEMP_sistelaan:         "2016-06-08",
	}

	lmarcRec := mustParseLmarc(lmarcDump)
	laanerRec := mustParseKeyVal(laanerDump)
	lnelRec := mustParseKeyVal(lnelDump)

	got := merge(lmarcRec, laanerRec, lnelRec)
	got.password = "" // bcrypt hash is different each time, so don't use in comparing
	if got != want {
		t.Errorf("got:\n%+v; want:\n%+v", got, want)
	}
}

func mustParseKeyVal(s string) map[string]string {
	dec := NewKVDecoder(bytes.NewBufferString(s))
	rec, err := dec.Decode()
	if err != nil {
		panic(err)
	}
	return rec
}

func mustParseLmarc(s string) marc.Record {
	enc := marc.NewDecoder(bytes.NewBufferString(s), marc.LineMARC)
	recs, err := enc.DecodeAll()
	if err != nil {
		panic(err)
	}
	if len(recs) != 1 {
		panic(fmt.Sprintf("got %d MARC records, want 1", len(recs)))
	}
	return recs[0]
}
