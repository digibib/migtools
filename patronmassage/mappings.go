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
  ("BHG","Barnegage","I","2999-12-31",\N,\N),
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

	staff = []string{"20171430", "626404", "655471", "460943", "246302", "470478", "87032", "20061870", "480100", "53666", "93540", "462760", "212835", "20017997", "880144", "889034", "900212", "209099", "20181475", "661111", "892061", "44657", "105495", "972928", "556318", "345216", "244167", "24119", "235", "6747", "830434", "20017677", "97062", "20075805", "332299", "20188883", "329441", "104220", "307251", "20006741", "897798", "95823", "86962", "600771", "20015035", "915743", "597307", "30", "843542", "212", "85210", "20074572", "6355", "828002", "860304", "54379", "830935", "1007", "1959", "453523", "52911", "20029116", "37228", "555007", "20051406", "34993", "1961", "226520", "900762", "48092", "45", "20184688", "1506", "486631", "845403", "20050693", "508383", "56", "34058", "634340", "20710", "512044", "20151423", "462445", "711987", "20101102", "20088594", "850085", "20075890", "20014372", "751922", "33551", "601043", "25684", "933840", "808708", "139010", "20176199", "20016889", "551720", "20051361", "240313", "178464", "20183097", "534774", "465019", "771012", "451201", "638027", "236553", "58359", "32928", "20069963", "827238", "769842", "941717", "294194", "862622", "185356", "914290", "20042964", "502106", "546", "20111498", "20015110", "614015", "20098494", "29", "40", "247039", "34520", "42", "772222", "20047476", "513414", "20016313", "587779", "32848", "64114", "757076", "32828", "378767", "934194", "561006", "496895", "809969", "781330", "859969", "515862", "20197902", "677371", "924328", "834900", "618224", "485644", "20187244", "777756", "696100", "615787", "787772", "20076200", "20169043", "162118", "275740", "879742", "370572", "3430", "21601", "3", "541508", "20202", "50411", "20020793", "1212", "695754", "34888", "570068", "374627", "4010", "665599", "20130846", "20079558", "61078", "858720", "20081567", "575437", "20000819", "940609", "20204635", "892357", "1955", "78873", "20194443", "862976", "518453", "20028601", "839604", "820791", "608943", "20130991", "787057", "234794", "20162415", "941170", "20034615", "20078355", "35102", "4005", "423629", "802005", "757420", "801686", "844443", "79583", "111465", "655519", "608665", "24679", "425688", "808825", "764191", "20142482", "753430"}

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
