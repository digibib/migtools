package main

import "github.com/boutros/marc"

var (
	itypesSQL = `
INSERT IGNORE INTO itemtypes
  (itemtype, description, notforloan)
VALUES
  ("0","Ikke til utlån","1"),
  ("1","Dagslån",""),
  ("14","14 dager",""),
  ("28","28 dager","");
`

	aValuesSQL = `
INSERT INTO authorised_values
  (category, authorised_value, lib)
VALUES
  ("WITHDRAWN","1","trukket tilbake"),
  ("DAMAGED","1","skadet"),
  ("LOST","1","tapt"),
  ("LOST","2","regnes som tapt"),
  ("LOST","3","tapt og erstattet"),
  ("LOST","4","ikke på plass"),
  ("LOST","5","påstått levert"),
  ("LOST","6","påstått ikke lånt"),
  ("LOST","7","borte i transport"),
  ("LOST","8","tapt, regning betalt"),
  ("LOST","9","vidvanke, registrert forsvunnet"),
  ("LOST","10","retur eieravdeling (ved import)"),
  ("LOST","11","til henteavdeling (ved import)"),
  ("NOT_LOAN","-1","i bestilling"),
  ("NOT_LOAN","2","ny"),
  ("NOT_LOAN","3","til internt bruk"),
  ("NOT_LOAN","4","til katalogisering"),
  ("NOT_LOAN","5","vurderes kassert"),
  ("NOT_LOAN","6","til retting"),
  ("NOT_LOAN","7","til innbinding"),
  ("RESTRICTED","1","begrenset tilgang"),
  ("RESTRICTED","2","referanseverk");
`
	issuesSQLtmp = `INSERT IGNORE INTO issues (borrowernumber, renewals, date_due, itemnumber)
SELECT borrowers.borrowernumber,
       {{.NumRes}},
       CONCAT('{{.DueDate}}', ' 23:59:00'),
       items.itemnumber
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
		"i": marc.SubField{Code: "1", Value: "4"},  // ikke på plass
		"p": marc.SubField{Code: "1", Value: "5"},  // påstått levert
		"l": marc.SubField{Code: "1", Value: "6"},  // påstått ikke lånt
		"V": marc.SubField{Code: "1", Value: "9"},  // på vidvanke
		"a": marc.SubField{Code: "1", Value: "10"}, // retur eieravdeling
		"y": marc.SubField{Code: "1", Value: "11"}, // til henteavdeling
	}

	// branchcode to label
	branchCodes = map[string]string{
		"dfb":    "Det Flerspråklige Bibliotek",
		"dfbs":   "Det Flerspråklige Bibliotek Referanse",
		"fbje":   "Bjerke",
		"fbjh":   "Bjerke,lokalhistorie",
		"fbji":   "Bjerke,innvandrerlitteratur",
		"fbjl":   "Bjørnholt læremidler",
		"fbjo":   "Bjørnholt",
		"fbju":   "Bjørnholt ungdomsskole",
		"fbli":   "vet ikke",
		"fbol":   "Bøler",
		"fboa":   "Bøler, Automater",
		"fdum":   "dummy",
		"ffur":   "Furuset",
		"ffua":   "Furuset, Automater",
		"fgam":   "Gamle Oslo",
		"fgaa":   "Gamle Oslo, Automater",
		"fgry":   "Grünerløkka",
		"fgyi":   "Grünerløkka innvandrerlitteratur",
		"fgra":   "Grünerløkka, Automater 1.etg",
		"fgrb":   "Grünerløkka, Automater 2.etg",
		"fhol":   "Holmlia",
		"fhoa":   "Holmlia, Automater",
		"flam":   "Lambertseter",
		"flaa":   "Lambertseter, Automater",
		"flan":   "Lambertseter, Nattautomat",
		"fmaj":   "Majorstua",
		"fmaa":   "Majorstua, Automater",
		"fnor":   "Nordtvet",
		"fnyd":   "Nydalen",
		"fnya":   "Nydalen, Automater",
		"fnyl":   "Nydalen, læremidler",
		"fopp":   "Oppsal",
		"fopa":   "Oppsal, Automater",
		"frik":   "Rikshospitalet",
		"frmm":   "Rommen",
		"frma":   "Rommen, Automater",
		"froa":   "Røa",
		"frob":   "Røa, Automater",
		"from":   "Romsås",
		"fsme":   "Smestad",
		"fsor":   "Sørkedalen nærbibliotek",
		"fsto":   "Stovner",
		"fsta":   "Stovner, Automater",
		"ftoi":   "Torshov innvandrerlitteratur",
		"ftoa":   "Torshov, Automater",
		"ftor":   "Torshov",
		"fxxx":   "Midlertidig filial X",
		"hbar":   "Barneavdelingen (Hovedutlånet)",
		"hbbr":   "Barneavdelingen spesialsamling",
		"hsko":   "Skoletjenesten",
		"hutl":   "Hovedbiblioteket",
		"hvkr":   "Katalogavdeling referanse",
		"hvlr":   "Stjernesamling, lesesalen",
		"hvmu":   "Musikkavdelingen (Hovedbiblioteket)",
		"hvma":   "Musikkavdelingen, Automater (Hovedbiblioteket)",
		"hvur":   "Spesialsamling, fjernlån",
		"hvua":   "Hovedbiblioteket, Automater",
		"idep":   "DFB,fjernlån",
		"info":   "Brosjyrelager",
		"mpmi":   "MappaMi",
		"ebib":   "eBokBib",
		"ukjent": "ukjent avdeling",
	}
)
