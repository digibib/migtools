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
  ("BHG","Barnehage","I","2999-12-31",\N,\N),
  ("BIB","Bibliotek","I","2999-12-31",\N,\N),
  ("I","Institusjon","I","2999-12-31",\N,\N),
  ("KL","Klasselåner","P","2999-12-31",\N,\N),
  ("MDL","Midlertidig bosatt","A","2999-12-31",\N,\N),
  ("PAS","Pasient","A", "2999-12-31",\N,\N),
  ("SKO","Grunnskole","I", "2999-12-31",\N,\N),
  ("V","Voksen","A","2999-12-31",\N,16),
  ("VGS","Videregående skole","I","2999-12-31",\N,\N);`

	fnrTemplSQL = `
INSERT IGNORE INTO borrower_attributes (borrowernumber, code, attribute)
SELECT borrowers.borrowernumber,
       'fnr',
       '{{.Fnr}}'
FROM borrowers
WHERE borrowers.userid = '{{.BibliofilBorrowerNr}}';
`
	categoryCodes = map[string]string{
		"v":   "V",
		"NB":  "BIB",
		"b":   "B",
		"u":   "V",
		"kl":  "KL",
		"NF":  "BIB",
		"pas": "PAS",
		"bhg": "BHG",
		"i":   "I",
		"EU":  "I",
		"sko": "SKO",
		"OV":  "I",
		"U03": "BIB",
		"G02": "BIB",
		"G12": "BIB",
		"G03": "BIB",
		"G16": "BIB",
		"G01": "BIB",
		"G18": "BIB",
		"G07": "BIB",
		"G11": "BIB",
		"B18": "BIB",
		"G04": "BIB",
		"B15": "BIB",
		"B12": "BIB",
		"G17": "BIB",
		"G15": "BIB",
		"V12": "BIB",
		"F03": "BIB",
		"B14": "BIB",
		"G05": "BIB",
		"G19": "BIB",
		"B11": "BIB",
		"U12": "BIB",
		"B19": "BIB",
		"B16": "BIB",
		"G06": "BIB",
		"G08": "BIB",
		"G09": "BIB",
		"B06": "BIB",
		"B05": "BIB",
		"B02": "BIB",
		"V11": "BIB",
		"V03": "BIB",
		"B04": "BIB",
		"B20": "BIB",
		"V02": "BIB",
		"U02": "BIB",
		"B17": "BIB",
		"B08": "BIB",
		"U16": "BIB",
		"B10": "BIB",
		"V18": "BIB",
		"G14": "BIB",
		"U11": "BIB",
		"B03": "BIB",
		"B01": "BIB",
		"V16": "BIB",
		"G10": "BIB",
		"V15": "BIB",
		"B07": "BIB",
		"B09": "BIB",
		"U18": "BIB",
		"V06": "BIB",
		"V04": "BIB",
		"V10": "BIB",
		"V05": "BIB",
		"V01": "BIB",
		"V08": "BIB",
		"U19": "BIB",
		"V19": "BIB",
		"V07": "BIB",
		"U04": "BIB",
		"V09": "BIB",
		"G20": "BIB",
		"V17": "BIB",
		"U20": "BIB",
		"U15": "BIB",
		"V14": "BIB",
		"U01": "BIB",
		"U14": "BIB",
		"U08": "BIB",
		"V20": "BIB",
		"U07": "BIB",
		"U06": "BIB",
		"U05": "BIB",
		"U17": "BIB",
		"U10": "BIB",
		"U09": "BIB",
		"F98": "BIB",
		"F11": "BIB",
		"F02": "BIB",
		"F16": "BIB",
		"F18": "BIB",
		"bkm": "V",
		"V21": "BIB",
		"U21": "BIB",
		"F20": "BIB",
		"F14": "BIB",
		"F07": "BIB",
		"U23": "BIB",
		"stl": "V",
		"F19": "BIB",
		"F06": "BIB",
		"F01": "BIB",
		"B21": "BIB",
	}

	branchesSQLtmpl = `
INSERT IGNORE INTO branches
  (branchcode, branchname)
VALUES
  {{$l := len . -}}
  {{range $i, $label := . -}}
  ("{{.Code}}","{{.Label}}"){{if eq (plus1 $i) $l}};{{else}},{{end}}
  {{end}}
`
	branchOldToNew = map[string]string{
		"fbjh": "fbje",
		"fbji": "fbje",
		"fbli": "fbol",
		"fgyi": "fgry",
		"fnti": "fnor",
		"fsti": "fsto",
		"ftoi": "ftor",
		"hbar": "hbib",
		"hbbr": "hbib",
		"hutl": "hbib",
		"hvkr": "hbib",
		"hvlr": "hbib",
		"hvmu": "hbib",
		"hvur": "hbib",
		"info": "hbib",
	}

	// branchcode to label
	branchCodes = map[string]string{
		"api":    "Internt API",
		"hbib":   "Hovedbiblioteket",
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

	msgPrefsInit = `
-- item due:
INSERT INTO borrower_message_preferences (borrowernumber, message_attribute_id)
  SELECT borrowernumber,1 FROM borrowers;

-- advance notice:
INSERT INTO borrower_message_preferences (borrowernumber, message_attribute_id, days_in_advance)
  SELECT borrowernumber,2,2 FROM borrowers;

-- hold filled:
INSERT INTO borrower_message_preferences (borrowernumber, message_attribute_id)
  SELECT borrowernumber,4 FROM borrowers;
`

	// Message for Item due
	msgDue = `
INSERT INTO borrower_message_transport_preferences (borrower_message_preference_id, message_transport_type)
SELECT borrower_message_preferences.borrower_message_preference_id,%q FROM borrower_message_preferences
  INNER JOIN borrowers ON borrower_message_preferences.borrowernumber = borrowers.borrowernumber
  AND message_attribute_id = 1
  WHERE borrowers.userid=%q;`

	// Message for advance notice
	msgNotice = `
INSERT INTO borrower_message_transport_preferences (borrower_message_preference_id, message_transport_type)
SELECT borrower_message_preferences.borrower_message_preference_id,%q FROM borrower_message_preferences
  INNER JOIN borrowers ON borrower_message_preferences.borrowernumber = borrowers.borrowernumber
  AND message_attribute_id = 2
  WHERE borrowers.userid=%q;`

	// Message for hold filled
	msgHold = `
INSERT INTO borrower_message_transport_preferences (borrower_message_preference_id, message_transport_type)
SELECT borrower_message_preferences.borrower_message_preference_id,%q FROM borrower_message_preferences
  INNER JOIN borrowers ON borrower_message_preferences.borrowernumber = borrowers.borrowernumber
  AND message_attribute_id = 4
  WHERE borrowers.userid=%q;
`
)
