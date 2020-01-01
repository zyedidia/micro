package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zyedidia/tcell"
)

func TestSimpleStringToStyle(t *testing.T) {
	s := StringToStyle("lightblue,magenta")

	fg, bg, _ := s.Decompose()

	assert.Equal(t, tcell.ColorBlue, fg)
	assert.Equal(t, tcell.ColorPurple, bg)
}

func TestAttributeStringToStyle(t *testing.T) {
	s := StringToStyle("bold cyan,brightcyan")

	fg, bg, attr := s.Decompose()

	assert.Equal(t, tcell.ColorTeal, fg)
	assert.Equal(t, tcell.ColorAqua, bg)
	assert.NotEqual(t, 0, attr&tcell.AttrBold)
}

func TestColor256StringToStyle(t *testing.T) {
	s := StringToStyle("128,60")

	fg, bg, _ := s.Decompose()

	assert.Equal(t, tcell.Color128, fg)
	assert.Equal(t, tcell.Color60, bg)
}

func TestColorHexStringToStyle(t *testing.T) {
	s := StringToStyle("#deadbe,#ef1234")

	fg, bg, _ := s.Decompose()

	assert.Equal(t, tcell.NewRGBColor(222, 173, 190), fg)
	assert.Equal(t, tcell.NewRGBColor(239, 18, 52), bg)
}

func TestColorschemeParser(t *testing.T) {
	testColorscheme := `color-link default "#F8F8F2,#282828"
color-link comment "#75715E,#282828"
# comment
color-link identifier "#66D9EF,#282828" #comment
color-link constant "#AE81FF,#282828"
color-link constant.string "#E6DB74,#282828"
color-link constant.string.char "#BDE6AD,#282828"`

	c, err := ParseColorscheme(testColorscheme)
	assert.Nil(t, err)

	fg, bg, _ := c["comment"].Decompose()
	assert.Equal(t, tcell.NewRGBColor(117, 113, 94), fg)
	assert.Equal(t, tcell.NewRGBColor(40, 40, 40), bg)
}
