package main

import (
	"bufio"
	"io"
	"strings"
	"time"

	"github.com/boutros/marc"
)

type patron struct {
	borrowernumber              int       // `borrowernumber` int(11) NOT NULL AUTO_INCREMENT,
	cardnumber                  string    // `cardnumber` varchar(16) DEFAULT NULL,
	surname                     string    // `surname` mediumtext NOT NULL,
	firstname                   string    // `firstname` text
	title                       string    //`title` mediumtext
	othernames                  string    //`othernames` mediumtext
	initials                    string    //`initials` text
	streetnumber                string    //`streetnumber` varchar(10) DEFAULT NULL,
	streettype                  string    //`streettype` varchar(50) DEFAULT NULL,
	address                     string    //`address` mediumtext NOT NULL,
	address2                    string    //`address2` text
	city                        string    //`city` mediumtext NOT NULL,
	state                       string    //`state` text
	zipcode                     string    //`zipcode` varchar(25) DEFAULT NULL,
	country                     string    //`country` text
	email                       string    //`email` mediumtext
	phone                       string    //`phone` text
	mobile                      string    //`mobile` varchar(50) DEFAULT NULL,
	fax                         string    //`fax` mediumtext
	emailpro                    string    //`emailpro` text
	phonepro                    string    //`phonepro` text
	B_streetnumber              string    //`B_streetnumber` varchar(10) DEFAULT NULL,
	B_streettype                string    //`B_streettype` varchar(50) DEFAULT NULL,
	B_address                   string    //`B_address` varchar(100) DEFAULT NULL,
	B_address2                  string    //`B_address2` text
	B_city                      string    //`B_city` mediumtext
	B_state                     string    //`B_state` text
	B_zipcode                   string    //`B_zipcode` varchar(25) DEFAULT NULL,
	B_country                   string    //`B_country` text
	B_email                     string    //`B_email` text
	B_phone                     string    //`B_phone` mediumtext
	dateofbirth                 time.Time //`dateofbirth` date DEFAULT NULL,
	branchcode                  string    //`branchcode` varchar(10) NOT NULL DEFAULT '',
	categorycode                string    //`categorycode` varchar(10) NOT NULL DEFAULT '',
	dateenrolled                string    //`dateenrolled` date DEFAULT NULL,
	dateexpiry                  string    //`dateexpiry` date DEFAULT NULL,
	gonenoaddress               bool      //`gonenoaddress` tinyint(1) DEFAULT NULL,
	lost                        bool      //`lost` tinyint(1) DEFAULT NULL,
	debarred                    time.Time //`debarred` date DEFAULT NULL,
	debarredcomment             string    //`debarredcomment` varchar(255) DEFAULT NULL,
	contactname                 string    //`contactname` mediumtext
	contactfirstname            string    //`contactfirstname` text
	contacttitle                string    //`contacttitle` text
	guarantorid                 int       //`guarantorid` int(11) DEFAULT NULL,
	borrowernotes               string    //`borrowernotes` mediumtext
	relationship                string    //`relationship` varchar(100) DEFAULT NULL,
	sex                         string    //`sex` varchar(1) DEFAULT NULL,
	password                    string    //`password` varchar(60) DEFAULT NULL,
	flags                       int       //`flags` int(11) DEFAULT NULL,
	usersid                     string    //`userid` varchar(75) DEFAULT NULL,
	opacnote                    string    //`opacnote` mediumtext
	contactnote                 string    //`contactnote` varchar(255) DEFAULT NULL,
	sort1                       string    //`sort1` varchar(80) DEFAULT NULL,
	sort2                       string    //`sort2` varchar(80) DEFAULT NULL,
	altcontactfirstname         string    //`altcontactfirstname` varchar(255) DEFAULT NULL,
	altcontactsurname           string    //`altcontactsurname` varchar(255) DEFAULT NULL,
	altcontactaddress1          string    //`altcontactaddress1` varchar(255) DEFAULT NULL,
	altcontactaddress2          string    //`altcontactaddress2` varchar(255) DEFAULT NULL,
	altcontactaddress3          string    //`altcontactaddress3` varchar(255) DEFAULT NULL,
	altcontactstate             string    //`altcontactstate` text
	altcontactzipcode           string    //`altcontactzipcode` varchar(50) DEFAULT NULL,
	altcontactcountry           string    //`altcontactcountry` text
	altcontactphone             string    //`altcontactphone` varchar(50) DEFAULT NULL,
	smsalertnumber              string    //`smsalertnumber` varchar(50) DEFAULT NULL,
	sms_provider                int       //`sms_provider_id` int(11) DEFAULT NULL,
	privacy                     int       //`privacy` int(11) NOT NULL DEFAULT '1',
	privacy_guarantor_checkouts int       //`privacy_guarantor_checkouts` tinyint(1) NOT NULL DEFAULT '0',
	updated_on                  time.Time //`updated_on` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
}

func parseKeyValRecord(in io.Reader) (map[string]string, error) {
	m := make(map[string]string, 24)
	r := bufio.NewReader(in)
	for {
		k, err := r.ReadString('|')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		v, err := r.ReadString('|')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// We have to check if we actually have reached the end of line,
		// because sometimes there are pipe-characters inside the value string.
		endChar, err := r.Peek(1)
		if err == nil {
			if string(endChar) != "\n" {
				vRest, err := r.ReadString('|')
				if err != nil {
					if err == io.EOF {
						break
					}
					return nil, err
				}
				v = v[:len(v)-1] + vRest
			}
		}

		// Cut at pipe-characters, and trim leading and trailing spaces and
		// any newlines inside string.
		m[strings.TrimSpace(k[0:len(k)-1])] = strings.Replace(
			strings.TrimSpace(v[0:len(v)-1]), "\n", "", 1)
	}
	return m, nil
}

func merge(lmarc marc.Record, laaner, lnel map[string]string) patron {
	res := patron{}
	return res
}
