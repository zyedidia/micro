package optionprovider

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var word = regexp.MustCompile(`[a-zA-Z]+\d*`)
var stopList = map[string]interface{}{
	"for":   false,
	"if":    false,
	"then":  false,
	"let":   false,
	"const": false,
	"when":  false,
	"while": false,
	"the":   false,
	"to":    false,
	"and":   false,
}
var maxSuggestions = 10

// Generic is an OptionProvider which provides options to the autocompletion system based on the
// words in the current buffer. It returns a delta of the start index if the start position needs
// to change.
func Generic(logger func(s string, values ...interface{}), buffer []byte, startOffset, currentOffset int) (options []Option, startDelta int, err error) {
	s := string(buffer)

	// Find the best matches.
	prefix := prefix(lastCharacters(s, currentOffset, 10))
	startDelta = currentOffset - startOffset - len(prefix)

	// Calculate common words.
	words := word.FindAllString(s, -1)

	// Count words, but remove common syntax and the prefix from the autocomplete list.
	counts := wordCounts(words, map[string]interface{}{prefix: false})
	orderedWords := orderByFrequencyDesc(counts)

	if len(prefix) > 0 {
		// We have a prefix, so write out prefix matches.
		for i := 0; i < len(orderedWords) && len(options) < maxSuggestions; i++ {
			if strings.HasPrefix(orderedWords[i], prefix) {
				options = append(options, New(orderedWords[i], ""))
			}
		}
		return
	}

	// Write out all matches.
	for i := 0; i < len(orderedWords) && len(options) < maxSuggestions; i++ {
		options = append(options, New(orderedWords[i], ""))
	}

	return
}

func lastCharacters(s string, end, charactersBefore int) string {
	if end == 0 {
		return ""
	}
	start := end - charactersBefore
	if start < 0 {
		return s[0:end]
	}
	return s[start:end]
}

func prefix(s string) string {
	text := []*unicode.RangeTable{unicode.Letter, unicode.Digit}
	last := 0
	for i, r := range s {
		if !unicode.IsOneOf(text, r) {
			last = i + 1
		}
	}
	return s[last:]
}

func wordCounts(words []string, additionalStopList map[string]interface{}) (counts map[string]int) {
	counts = make(map[string]int)
	for _, w := range words {
		if !inAnyStopList(w, stopList, additionalStopList) {
			counts[w]++
		}
	}
	return
}

func inAnyStopList(w string, lists ...map[string]interface{}) bool {
	for _, l := range lists {
		if _, inStopList := l[w]; inStopList {
			return true
		}
	}
	return false
}

func orderByFrequencyDesc(counts map[string]int) []string {
	cs := NewCountSorter(counts)
	sort.Sort(cs)
	return cs.Keys
}

// CountSorter provides a way of sorting map[string]int values (most frequent first).
type CountSorter struct {
	Data map[string]int
	Keys []string
}

// NewCountSorter supports the sort.Data interface. The Keys field is populated, and the
// struct can then be sorted. The Keys field then contains the keys, ordered by the count.
func NewCountSorter(data map[string]int) *CountSorter {
	cs := new(CountSorter)
	cs.Data = make(map[string]int)
	cs.Keys = make([]string, len(data))
	var i int
	for k, v := range data {
		cs.Keys[i] = k
		cs.Data[k] = v
		i++
	}
	return cs
}

func (cs *CountSorter) Len() int { return len(cs.Data) }
func (cs *CountSorter) Less(i, j int) bool {
	countI := cs.Data[cs.Keys[i]]
	countJ := cs.Data[cs.Keys[j]]
	if countI == countJ {
		// Alphabetic sorting.
		// Go's sorting rules mean that capital letters come before lowercase.
		return cs.Keys[i] < cs.Keys[j]
	}
	// Sort by largest count first.
	return countI > countJ
}
func (cs *CountSorter) Swap(i, j int) { cs.Keys[i], cs.Keys[j] = cs.Keys[j], cs.Keys[i] }
