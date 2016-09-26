package main

import "github.com/boutros/marc"

var (
	itypesSQL = `
INSERT IGNORE INTO itemtypes
  (itemtype, description, notforloan)
VALUES
  ("0","Ikke til utlån","1"),
  ("DAGSLAAN","Dagslån",""),
  ("UKESLAAN","Hurtiglån (7 dager)",""),
  ("TOUKESLAAN","Hurtiglån (14 dager)",""),
  ("SPRAAKKURS","Språkkurs",""),
  ("LYDBOK","Lydbok",""),
  ("MUSIKK","Musikkopptak",""),
  ("SPILL","Spill",""),
  ("FILM","Film",""),
  ("EBOK","E-bok",""),
  ("PERIODIKA","Periodika",""),
  ("BOK","Bok",""),
  ("NOTER","Noter",""),
  ("KART","Kart",""),
  ("REALIA","Realia",""),
  ("UKJENT", "Ukjent", ""),
  ("","","");
`

	issuesSQLtmp = `INSERT IGNORE INTO issues (borrowernumber, renewals, date_due, itemnumber, branchcode)
SELECT borrowers.borrowernumber,
       {{.NumRes}},
       CONCAT('{{.DueDate}}', ' 23:59:00'),
       items.itemnumber,
       '{{.Branch}}'
FROM borrowers
INNER JOIN items ON items.barcode = '{{.Barcode}}'
WHERE borrowers.userid = '{{.BibliofilBorrowerNr}}';
`

	branchesSQLtmpl = `
INSERT IGNORE INTO branches
  (branchcode, branchname)
VALUES
  {{$l := len . -}}
  {{range $i, $label := . -}}
  ("{{.Code}}","{{.Label}}"){{if eq (plus1 $i) $l}};{{else}},{{end}}
  {{end}}
`

	statusCodes = map[string]marc.SubField{
		// NOT_LOAN values: (negative value => can be reseved):

		"e": marc.SubField{Code: "7", Value: "-1"}, // i bestilling
		"n": marc.SubField{Code: "7", Value: "2"},  // ny
		"c": marc.SubField{Code: "7", Value: "3"},  // til internt bruk
		"k": marc.SubField{Code: "7", Value: "4"},  // til katalogisering
		"v": marc.SubField{Code: "7", Value: "5"},  // vurderes kassert
		"q": marc.SubField{Code: "7", Value: "6"},  // retting
		"b": marc.SubField{Code: "7", Value: "7"},  // til innbinding

		// LOST values:
		"t": marc.SubField{Code: "1", Value: "1"},  // tapt
		"S": marc.SubField{Code: "1", Value: "8"},  // tapt, regning betalt
		"r": marc.SubField{Code: "1", Value: "12"}, // regning sendt
		"i": marc.SubField{Code: "1", Value: "4"},  // ikke på plass
		"p": marc.SubField{Code: "1", Value: "5"},  // påstått levert
		"l": marc.SubField{Code: "1", Value: "6"},  // påstått ikke lånt
		"V": marc.SubField{Code: "1", Value: "9"},  // på vidvanke
		"a": marc.SubField{Code: "1", Value: "10"}, // retur eieravdeling
		"y": marc.SubField{Code: "1", Value: "11"}, // til henteavdeling
	}

	branchOldToNew = map[string]string{
		"fbjh": "fbje",
		"fbji": "fbje",
		"fbli": "fbol",
		"fgyi": "fgry",
		"fnti": "fnor",
		"fsti": "fsto",
		"ftoi": "ftor",
		"hbbr": "hbar",
		"hvkr": "hutl",
		"hvlr": "hutl",
		"hvur": "hutl",
		"info": "hutl",
		"fstl": "fsto",
		// Automat-avdelinger:
		"fboa": "fbol",
		"ffua": "ffur",
		"fgaa": "fgam",
		"fgra": "fgry",
		"fgrb": "frgy",
		"fhoa": "fhol",
		"flaa": "flam",
		"flan": "flam",
		"fmaa": "fmaj",
		"fnya": "fnyd",
		"fopa": "fopp",
		"frma": "frmm",
		"frob": "froa",
		"ftoa": "ftor",
		"hvma": "hvmu",
		"hvua": "hutl",
	}

	// branchcode to label
	branchCodes = map[string]string{
		"api":    "Internt API",
		"hutl":   "Hovedbiblioteket, voksen",
		"hbar":   "Hovedbiblioteket, barn",
		"hvmu":   "Hovedbiblioteket, musikk",
		"fbje":   "Bjerke",
		"fbjo":   "Bjørnholt",
		"fbol":   "Bøler",
		"ffur":   "Furuset",
		"fgab":   "Biblo Tøyen",
		"fgam":   "Tøyen",
		"fgry":   "Grünerløkka",
		"fhol":   "Holmlia",
		"flam":   "Lambertseter",
		"fmaj":   "Majorstuen",
		"fnor":   "Nordtvet",
		"fnyd":   "Nydalen",
		"fopp":   "Oppsal",
		"frik":   "Rikshospitalet",
		"frmm":   "Rommen",
		"froa":   "Røa",
		"from":   "Romsås",
		"fsme":   "Smestad",
		"fsto":   "Stovner",
		"ftor":   "Torshov",
		"hsko":   "Skoletjenesten",
		"ukjent": "Ukjent avdeling",
	}
)
