package main

var (
	categoriesSQL = `
INSERT IGNORE INTO categories
  (categorycode, description, category_type, enrolmentperioddate, upperagelimit, dateofbirthrequired )
VALUES
  ("ADMIN","Administrator","S","2999-12-31",\N,\N),
  ("ANS","Ansatt","S","2999-12-31",\N,\N),
  ("API","API user","S","2999-12-31",\N,\N),
  ("AUTO","Automat","S","2999-12-31",\N,\N),
  ("B","Barn","C","2999-12-31",15,0),
  ("BGH","Barnegage","I","2999-12-31",\N,\N),
  ("BIB","Bibliotek","I","2999-12-31",\N,\N),
  ("I","Institusjon","I","2999-12-31",\N,\N),
  ("KL","Klasselåner","P","2999-12-31",\N,\N),
  ("MDL","Midlertidig bosatt","A","2999-12-31",\N,\N),
  ("PAS","Pasient","A", "2999-12-31",\N,\N),
  ("SKO","Grunnskole","I", "2999-12-31",\N,\N),
  ("V","Voksen","A","2999-12-31",\N,16),
  ("VGS","Videregående skole","I","2999-12-31",\N,\N);`

	branchesSQLtmpl = `
INSERT IGNORE INTO branches
  (branchcode, branchname)
VALUES
  {{$l := len . -}}
  {{range $i, $label := . -}}
  ("{{.Code}}","{{.Label}}"){{if eq (plus1 $i) $l}};{{else}},{{end}}
  {{end}}
`

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
