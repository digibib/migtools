package main

import (
	"bytes"
	"strings"
	"time"
	"unicode"

	"github.com/boutros/marc"
)

const (
	noDateFormat    = "02/01/2006"
	mysqlDateFormat = "2006-01-02"
)

type patron struct {
	cardnumber     string // `cardnumber` varchar(16) DEFAULT NULL,
	userid         string // `userid` varchar(75) DEFAULT NULL,
	surname        string // `surname` mediumtext NOT NULL,
	firstname      string // `firstname` text
	address        string // `address` mediumtext NOT NULL,
	address2       string // `address2` text
	city           string // `city` mediumtext NOT NULL,
	zipcode        string // `zipcode` varchar(25) DEFAULT NULL,
	country        string // `country` text
	email          string // `email` mediumtext
	phone          string // `phone` text
	smsalertnumber string // `smsalertnumber` varchar(50) DEFAULT NULL,
	dateofbirth    string // `dateofbirth` date DEFAULT NULL,
	branchcode     string // `branchcode` varchar(10) NOT NULL DEFAULT '',
	categorycode   string // `categorycode` varchar(10) NOT NULL DEFAULT '',
	dateenrolled   string // `dateenrolled` date DEFAULT NULL,
	dateexpiry     string // `dateexpiry` date DEFAULT NULL,
	gonenoaddress  bool   // `gonenoaddress` tinyint(1) DEFAULT NULL,
	lost           bool   // `lost` tinyint(1) DEFAULT NULL,
	borrowernotes  string // `borrowernotes` mediumtext
	sex            string // `sex` varchar(1) DEFAULT NULL,
	password       string // `password` varchar(60) DEFAULT NULL,
	privacy        int    // `privacy` int(11) NOT NULL DEFAULT '1',

	// Temporary variables that have no matching column in the borrowers table,
	// but we need the information for further processing or populating borrower-connected tables.
	TEMP_sistelaan         string
	TEMP_personnr          string
	TEMP_nl                bool
	TEMP_hjemmebibnr       string
	TEMP_res_transport     string
	TEMP_pur_transport     string
	TEMP_fvarsel_transport string
}

// splitZipCity splits string into zip code and city. If there is no
// match in zipcode, the whole input string will be returned as the second return value.
func splitZipCity(s string) (string, string) {
	i := 0
	for ; i < len(s); i++ {
		if !('0' <= s[i] && s[i] <= '9') {
			break
		}
		if i == 4 {
			// zip code is max 4 digits
			break
		}
	}
	return s[:i], strings.TrimSpace(s[i:])
}

// firstSub returns the first value of the code in subfields
func firstSub(s marc.SubFields, code string) string {
	for _, f := range s {
		if f.Code == code {
			return f.Value
		}
	}
	return ""
}

// onlyDigits strip all characters from string except digits and '+' sign
func onlyDigits(s string) string {
	var r bytes.Buffer
	for _, c := range s {
		if unicode.IsDigit(c) || c == '+' {
			r.WriteRune(c)
		}
	}
	return r.String()
}

func merge(lmarc *marc.Record, laaner, lnel map[string]string) patron {
	// defaults:
	p := patron{
		privacy:    1,
		dateexpiry: "2099-01-01",
		branchcode: "hutl",
	}

	// 1) information from laaner
	i := strings.Index(laaner["ln_navn"], ",")
	if i != -1 {
		p.surname = laaner["ln_navn"][:i]
		p.firstname = strings.TrimSpace(laaner["ln_navn"][i+1:])
	} else {
		p.surname = strings.TrimSpace(laaner["ln_navn"])
	}

	dob, err := time.Parse(noDateFormat, laaner["ln_foedt"])
	if err == nil {
		p.dateofbirth = dob.Format(mysqlDateFormat)
	}

	p.address = laaner["ln_adr1"]
	p.address2 = laaner["ln_adr2"]
	p.zipcode, p.city = splitZipCity(laaner["ln_post"])
	p.country = laaner["ln_land"]
	p.phone = laaner["ln_tlf"]
	p.categorycode = laaner["ln_kat"]

	switch laaner["ln_kjoenn"] {
	case "k":
		p.sex = "F"
	case "m":
		p.sex = "M"
	}

	// we store bibliofil lånenr temporarily as userid, so that
	// we can match loans etc on this. Later to be changed to be the cardnumber.
	p.userid = laaner["ln_nr"]

	p.borrowernotes = laaner["ln_melding"]

	if laaner["ln_sistelaan"] != "00/00/0000" {
		d, err := time.Parse(noDateFormat, laaner["ln_sistelaan"])
		if err == nil {
			p.TEMP_sistelaan = d.Format(mysqlDateFormat)
		}
	}

	if laaner["ln_kortdato"] != "00/00/0000" {
		d, err := time.Parse(noDateFormat, laaner["ln_kortdato"])
		if err == nil {
			p.dateenrolled = d.Format(mysqlDateFormat)
		}
	}

	if strings.Contains(laaner["ln_obs"]+laaner["ln_friobs"], "m") {
		p.lost = true
	}

	if strings.Contains(laaner["ln_obs"]+laaner["ln_friobs"], "f") {
		p.gonenoaddress = true
	}

	// 2) information from lmarc

	for _, f := range lmarc.DataFields {
		switch f.Tag {
		case "105":
			// TODO foresatte = p.contactname ?
		case "140":
			bCode := firstSub(f.SubFields, "a")
			if bCode != "" && len(bCode) <= 4 && len(bCode) >= 3 {
				// filter out bad data, accepting only 3 or 4 character labels
				p.branchcode = bCode
			}
			bCode = firstSub(f.SubFields, "b")
			if bCode != "" && len(bCode) <= 4 && len(bCode) >= 3 {
				// 140$b = foretrukken henteavdeling, ant. mer oppdatert enn 140$a,
				// som sier hvor låneren ble registrert.
				p.branchcode = bCode
			}
		case "150":
			// TODO melding = p.borrowernotes?
			// NB feltet er repeterbart
		case "190":
			if v := firstSub(f.SubFields, "a"); len(v) >= 11 {
				p.TEMP_personnr = v
			}
		case "200":
			if v := firstSub(f.SubFields, "s"); v == "1" {
				p.gonenoaddress = true
			}
		case "240":
			// telefonnr (repeterbart felt)
			// $c = fax|jobb|mobil|mobilsms
			v := onlyDigits(firstSub(f.SubFields, "a"))
			switch firstSub(f.SubFields, "c") {
			case "jobb":
				if p.phone == "" {
					p.phone = v
				}
			case "mobil", "mobilsms":
				p.smsalertnumber = v
			}
		case "261":
		case "270": // transporttype reserveringsbrev
			p.TEMP_res_transport = strings.ToLower(firstSub(f.SubFields, "a"))
		case "271": // transporttype purring
			p.TEMP_pur_transport = strings.ToLower(firstSub(f.SubFields, "a"))
		case "272": // transporttype forhåndsvarsel
			p.TEMP_fvarsel_transport = strings.ToLower(firstSub(f.SubFields, "a"))
		case "300": // Lagre historikk
			if firstSub(f.SubFields, "a") == "1" {
				// 0 = forever, 1 = default, 2 = never
				p.privacy = 0
			} else {
				p.privacy = 2
			}
		case "600": // Nasjonalt lånenummer
			p.cardnumber = firstSub(f.SubFields, "a")
			// 600$k = 1 hvis tilknyttet NL, 0 hvis ikke
			if v := firstSub(f.SubFields, "k"); v != "" {
				p.TEMP_nl = true
			}
		case "606":
			if v := firstSub(f.SubFields, "b"); len(v) >= 11 {
				p.TEMP_personnr = v
			}
		}
	}

	// 3) information from lnel
	p.email = strings.TrimSpace(lnel["lnel_epost"])

	return p
}
