package main

import (
	"bytes"
	"testing"

	"github.com/knakk/kbp/marc"
)

func mustDecode(input string) *marc.Record {
	l := 64
	if len(input) < 64 {
		l = len(input)
	}

	format := marc.DetectFormat([]byte(input[:l]))
	switch format {
	case marc.MARC, marc.LineMARC, marc.MARCXML:
		break
	default:
		panic("mustDeocde: Unknown MARC format")
	}

	// Decode whole input stream
	dec := marc.NewDecoder(bytes.NewBufferString(input), format)
	recs, err := dec.DecodeAll()
	if err != nil {
		panic("mustDecode: " + err.Error())
	}

	// We only want one record
	if len(recs) != 1 {
		panic("mustDecode: expected one record")
	}
	return recs[0]
}

var tests = []struct {
	input string
	want  string
}{
	{
		`
*000     n
*0011913015
*00842536                 a          10bul
*015  $a10561242$bBibliofilID
*015  $bDFB
*019  $dR$bl$s42$aaa$bda,db
*020  $a978-619-161-070-9
*041  $hnor
*0820 $223/nor$zh
*090  $cBUL$dFla
*099  $axbulv160731
*10010$aFlatland, Helga$d1984-$jn.$32043547900
*24010$aBli hvis du kan. Reis hvis du må
*24510$aOstani, ako mozhjesh. Zamini, ako trjabva.$cHelga Flatlan ; [overs. av] Rostislav Petrov
*24510$aОстани, ако можеш. Замини, ако трябва.$cХелга Флатлан ; [overs. av] Ростислав Петров$9bul
*260  $aSofia$bPerseus$c2015
*300  $a224 s.
*520  $aEn oppvekstroman om tre barndomsvenner som verver seg til de norske styrkene i Afghanistan. De kommer aldri hjem. Hvorfor dro de? Hva skjer med de som er igjen?
*574  $aOriginaltittel: Bli hvis du kan. Reis hvis du må
*690  $aBygdesamfunn$xFortellinger$32047905900
*690  $aDøden$xFortellinger$32047013900
*690  $aOppvekst$32047012400
*690  $aSorg$xFortellinger$32046219400
*850  $aDEICHM$sn
^`,
		`
*000
*008                                 1
*020  $a978-619-161-070-9
*041  $abul
*100  $aFlatland, Helga
*245  $aOstani, ako mozhjesh. Zamini, ako trjabva.
*260  $aSofia$bPerseus$c2015
*338  $aVinylplate
*338  $aKassett
*385  $aVoksne
*385  $a0-2 år
*521  $a42
*650  $aBygdesamfunn
*650  $aDøden
*650  $aOppvekst
*650  $aFortellinger
*650  $aSorg
^`,
	},
}

func TestTransformation(t *testing.T) {
	for _, tt := range tests {
		from := mustDecode(tt.input)
		want := mustDecode(tt.want)
		got := Transform(from)
		if !got.Eq(want) {
			t.Errorf("got:\n%v\nwant:\n%v\n", got, want)
		}
	}
}
