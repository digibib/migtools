package main

import (
	"github.com/knakk/kbp/marc"
	"github.com/knakk/kbp/marc/normarc"
)

func copyCtrlFieldPos(from, to *marc.Record, tag marc.ControlTag, pos ...int) {
	f := marc.NewControlField(tag)
	for _, p := range pos {
		if cf, ok := from.ControlField(tag); ok {
			f.SetPos(p, cf.GetPos(p, 1))
		}
	}
	to.AddControlField(f)
}

func copyDataFieldSubfields(from, to *marc.Record, tag marc.DataTag, subfields ...rune) {
	copyDataFieldSubfieldsTo(from, to, tag, tag, subfields...)
}

func copyDataFieldSubfieldsTo(from, to *marc.Record, fromTag, toTag marc.DataTag, subfields ...rune) {
	for _, df := range from.DataFields(fromTag) {
		found := false
		f := marc.NewDataField(toTag)
		for _, code := range subfields {
			for _, v := range df.Subfield(code) {
				f.Add(code, v)
				found = true
			}
		}
		if found {
			to.AddDataField(f)
		}
	}
}

func Transform(from *marc.Record) *marc.Record {
	to := marc.NewRecord()

	// Fag/fiksjon *008 pos. 33 (Fag: 0, Fiksjon: 1)
	copyCtrlFieldPos(from, to, marc.Tag008, normarc.PosLitterærForm)

	// ISBN *020 $a, tag repeteres hvis flere ISBN
	copyDataFieldSubfields(from, to, marc.Tag020, 'a')

	// Språk *041 $a (ISO-kode, 3 tegn), delfelt repeteres hvis flere språk
	df := marc.NewDataField(marc.Tag041)
	if cf, ok := from.ControlField(marc.Tag008); ok {
		lang := cf.GetPos(normarc.PosSpråk, 3)
		if lang != "" {
			df.Add('a', lang)
		}
	}
	if from041, ok := from.DataField(marc.Tag041); ok {
		for _, v := range from041.Subfield('a') {
			df.Add('a', v)
		}
	}

	if len(df.Subfield('a')) > 0 {
		to.AddDataField(df)
	}

	//	Oppstilling, kategori	 	*090 $a
	//	Oppstilling, format			*090 $b
	//	Oppstilling, deweynummer	*090 $c
	//	Oppstilling, plassering		*090 $d

	// Navn på hovedinnførsel (person) 		*100 $a
	copyDataFieldSubfields(from, to, marc.Tag100, 'a')

	// Navn på hovedinnførsel (korporasjon) 110 $a
	copyDataFieldSubfields(from, to, marc.Tag110, 'a')

	// Hovedtittel 		*245 $a
	// Undertittel	 	*245 $b
	// Delnummer 		*245 $n
	// Deltittel 		*245 $p
	titleSubfields := []rune{'a', 'b', 'n', 'p'}
	ds := from.DataFields(marc.Tag245)
	if len(ds) == 1 {
		copyDataFieldSubfields(from, to, marc.Tag245, titleSubfields...)
	} else {
		// Poste har flere 245-felt (DFB)
		for _, d := range ds {
			if len(d.Subfield('9')) == 0 {
				// Hvis $9 er satt, vet vi at det er den utranskriberte tittelen.
				// Vi vil ha den transkriberte.
				for _, c := range titleSubfields {
					if sf := d.Subfield(c); len(sf) > 0 {
						to.AddDataField(marc.NewDataField(marc.Tag245).Add(c, sf[0]))
					}
				}
			}
		}
	}

	// Utgivelsessted	*260 $a (label)
	// Forlagsnavn 		*260 $b (label)
	// Utgivelsesår 	*260 $c
	copyDataFieldSubfields(from, to, marc.Tag260, 'a', 'b', 'c')

	// Verkstype 	*336 $a (label), tag repeteres hvis flere verkstyper
	// Medietype 	*337 $a (label), tag repeteres hvis flere medietyper
	// Format	*338 $a (label), tag repeteres hvis flere formater
	// Målgruppe 	*385 $a (label), tag repeteres hvis flere målgrupper
	// Tilrettelegging for bestemte brukergrupper	*385 $a (label), tag repeteres hvis flere grupper

	// Aldersgrense		*521 $a
	if df, ok := from.DataField(marc.Tag019); ok {
		if sf := df.Subfield('s'); len(sf) == 1 {
			to.AddDataField(marc.NewDataField(marc.Tag521).Add('a', sf[0]))
		}
	}

	// Emne		*650 $a (label), tag repeteres hvis flere emner
	copyDataFieldSubfieldsTo(from, to, marc.Tag690, marc.Tag650, 'a', 'x')

	// Emne med specification *650 $a (label) + $q (specification), tag repeteres hvis flere emner
	// Sjanger	*655 $a (label), tag repeteres hvis flere sjangrer
	// Litterær form	*655 $a (label), tag repeteres hvis flere sjangrer

	return to
}
