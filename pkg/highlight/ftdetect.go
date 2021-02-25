package highlight

import "regexp"

// MatchFiletype will use the list of syntax definitions provided and the filename and first line of the file
// to determine the filetype of the file
// It will return the corresponding syntax definition for the filetype
func MatchFiletype(ftdetect [2]*regexp.Regexp, filename string, firstLine []byte) bool {
	if ftdetect[0] != nil && ftdetect[0].MatchString(filename) {
		return true
	}

	if ftdetect[1] != nil {
		return ftdetect[1].Match(firstLine)
	}

	return false
}
