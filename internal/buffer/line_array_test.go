package buffer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var unicode_txt = `An preost wes on leoden, Laȝamon was ihoten
He wes Leovenaðes sone -- liðe him be Drihten.
He wonede at Ernleȝe at æðelen are chirechen,
Uppen Sevarne staþe, sel þar him þuhte,
Onfest Radestone, þer he bock radde.`

var la *LineArray

func init() {
	reader := strings.NewReader(unicode_txt)
	la = NewLineArray(uint64(len(unicode_txt)), FFAuto, reader)
}

func TestSplit(t *testing.T) {
	la.insert(Loc{17, 1}, []byte{'\n'})
	assert.Equal(t, len(la.lines), 6)
	sub1 := la.Substr(Loc{0, 1}, Loc{17, 1})
	sub2 := la.Substr(Loc{0, 2}, Loc{30, 2})

	assert.Equal(t, []byte("He wes Leovenaðes"), sub1)
	assert.Equal(t, []byte(" sone -- liðe him be Drihten."), sub2)
}

func TestJoin(t *testing.T) {
	la.remove(Loc{47, 1}, Loc{0, 2})
	assert.Equal(t, len(la.lines), 5)
	sub := la.Substr(Loc{0, 1}, Loc{47, 1})
	bytes := la.Bytes()

	assert.Equal(t, []byte("He wes Leovenaðes sone -- liðe him be Drihten."), sub)
	assert.Equal(t, unicode_txt, string(bytes))
}

func TestInsert(t *testing.T) {
	la.insert(Loc{20, 3}, []byte(" foobar"))
	sub1 := la.Substr(Loc{0, 3}, Loc{50, 3})

	assert.Equal(t, []byte("Uppen Sevarne staþe, foobar sel þar him þuhte,"), sub1)

	la.insert(Loc{25, 2}, []byte("H̼̥̯͇͙̕͘͞e̸̦̞̠̣̰͙̼̥̦̼̖̬͕͕̰̯̫͇̕ĺ̜̠̩̯̯͙̼̭̠͕̮̞͜l̶͓̫̞̮͈͞ͅo̸͔͙̳̠͈̮̼̳͙̥̲͜͠"))

	sub2 := la.Substr(Loc{0, 2}, Loc{60, 2})
	assert.Equal(t, []byte("He wonede at Ernleȝe at æH̼̥̯͇͙̕͘͞e̸̦̞̠̣̰͙̼̥̦̼̖̬͕͕̰̯̫͇̕ĺ̜̠̩̯̯͙̼̭̠͕̮̞͜l̶͓̫̞̮͈͞ͅo̸͔͙̳̠͈̮̼̳͙̥̲͜͠ðelen are chirechen,"), sub2)
}

func TestRemove(t *testing.T) {
	la.remove(Loc{20, 3}, Loc{27, 3})
	la.remove(Loc{25, 2}, Loc{30, 2})

	bytes := la.Bytes()
	assert.Equal(t, unicode_txt, string(bytes))
}
