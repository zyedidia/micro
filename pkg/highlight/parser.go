package highlight

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v2"
)

// A Group represents a syntax group
type Group uint8

// Groups contains all of the groups that are defined
// You can access them in the map via their string name
var Groups map[string]Group
var numGroups Group

// String returns the group name attached to the specific group
func (g Group) String() string {
	for k, v := range Groups {
		if v == g {
			return k
		}
	}
	return ""
}

// A Def is a full syntax definition for a language
// It has a filetype, information about how to detect the filetype based
// on filename or header (the first line of the file)
// Then it has the rules which define how to highlight the file
type Def struct {
	*Header

	rules *rules
}

type Header struct {
	FileType string
	FtDetect [2]*regexp.Regexp
}

type HeaderYaml struct {
	FileType string `yaml:"filetype"`
	Detect   struct {
		FNameRgx  string `yaml:"filename"`
		HeaderRgx string `yaml:"header"`
	} `yaml:"detect"`
}

type File struct {
	FileType string

	yamlSrc map[interface{}]interface{}
}

// A Pattern is one simple syntax rule
// It has a group that the rule belongs to, as well as
// the regular expression to match the pattern
type pattern struct {
	group Group
	regex *regexp.Regexp
}

// rules defines which patterns and regions can be used to highlight
// a filetype
type rules struct {
	regions  []*region
	patterns []*pattern
	includes []string
}

// A region is a highlighted region (such as a multiline comment, or a string)
// It belongs to a group, and has start and end regular expressions
// A region also has rules of its own that only apply when matching inside the
// region and also rules from the above region do not match inside this region
// Note that a region may contain more regions
type region struct {
	group      Group
	limitGroup Group
	parent     *region
	start      *regexp.Regexp
	end        *regexp.Regexp
	skip       *regexp.Regexp
	rules      *rules
}

func init() {
	Groups = make(map[string]Group)
}

// MakeHeader takes a header (.hdr file) file and parses the header
// Header files make parsing more efficient when you only want to compute
// on the headers of syntax files
// A yaml file might take ~400us to parse while a header file only takes ~20us
func MakeHeader(data []byte) (*Header, error) {
	lines := bytes.Split(data, []byte{'\n'})
	if len(lines) < 3 {
		return nil, errors.New("Header file has incorrect format")
	}
	header := new(Header)
	var err error
	header.FileType = string(lines[0])
	fnameRgx := string(lines[1])
	headerRgx := string(lines[2])

	if fnameRgx != "" {
		header.FtDetect[0], err = regexp.Compile(fnameRgx)
	}
	if headerRgx != "" {
		header.FtDetect[1], err = regexp.Compile(headerRgx)
	}

	if err != nil {
		return nil, err
	}

	return header, nil
}

// MakeHeaderYaml takes a yaml spec for a syntax file and parses the
// header
func MakeHeaderYaml(data []byte) (*Header, error) {
	var hdrYaml HeaderYaml
	err := yaml.Unmarshal(data, &hdrYaml)
	if err != nil {
		return nil, err
	}

	header := new(Header)
	header.FileType = hdrYaml.FileType

	if hdrYaml.Detect.FNameRgx != "" {
		header.FtDetect[0], err = regexp.Compile(hdrYaml.Detect.FNameRgx)
	}
	if hdrYaml.Detect.HeaderRgx != "" {
		header.FtDetect[1], err = regexp.Compile(hdrYaml.Detect.HeaderRgx)
	}

	if err != nil {
		return nil, err
	}

	return header, nil
}

func ParseFile(input []byte) (f *File, err error) {
	// This is just so if we have an error, we can exit cleanly and return the parse error to the user
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()

	var rules map[interface{}]interface{}
	if err = yaml.Unmarshal(input, &rules); err != nil {
		return nil, err
	}

	f = new(File)
	f.yamlSrc = rules

	for k, v := range rules {
		if k == "filetype" {
			filetype := v.(string)

			f.FileType = filetype
			break
		}
	}

	return f, err
}

// ParseDef parses an input syntax file into a highlight Def
func ParseDef(f *File, header *Header) (s *Def, err error) {
	// This is just so if we have an error, we can exit cleanly and return the parse error to the user
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()

	rules := f.yamlSrc

	s = new(Def)
	s.Header = header

	for k, v := range rules {
		if k == "rules" {
			inputRules := v.([]interface{})

			rules, err := parseRules(inputRules, nil)
			if err != nil {
				return nil, err
			}

			s.rules = rules
		}
	}

	return s, err
}

// HasIncludes returns whether this syntax def has any include statements
func HasIncludes(d *Def) bool {
	hasIncludes := len(d.rules.includes) > 0
	for _, r := range d.rules.regions {
		hasIncludes = hasIncludes || hasIncludesInRegion(r)
	}
	return hasIncludes
}

func hasIncludesInRegion(region *region) bool {
	hasIncludes := len(region.rules.includes) > 0
	for _, r := range region.rules.regions {
		hasIncludes = hasIncludes || hasIncludesInRegion(r)
	}
	return hasIncludes
}

// GetIncludes returns a list of filetypes that are included by this syntax def
func GetIncludes(d *Def) []string {
	includes := d.rules.includes
	for _, r := range d.rules.regions {
		includes = append(includes, getIncludesInRegion(r)...)
	}
	return includes
}

func getIncludesInRegion(region *region) []string {
	includes := region.rules.includes
	for _, r := range region.rules.regions {
		includes = append(includes, getIncludesInRegion(r)...)
	}
	return includes
}

// ResolveIncludes will sort out the rules for including other filetypes
// You should call this after parsing all the Defs
func ResolveIncludes(def *Def, files []*File) {
	resolveIncludesInDef(files, def)
}

func resolveIncludesInDef(files []*File, d *Def) {
	for _, lang := range d.rules.includes {
		for _, searchFile := range files {
			if lang == searchFile.FileType {
				searchDef, _ := ParseDef(searchFile, nil)
				d.rules.patterns = append(d.rules.patterns, searchDef.rules.patterns...)
				d.rules.regions = append(d.rules.regions, searchDef.rules.regions...)
			}
		}
	}
	for _, r := range d.rules.regions {
		resolveIncludesInRegion(files, r)
		r.parent = nil
	}
}

func resolveIncludesInRegion(files []*File, region *region) {
	for _, lang := range region.rules.includes {
		for _, searchFile := range files {
			if lang == searchFile.FileType {
				searchDef, _ := ParseDef(searchFile, nil)
				region.rules.patterns = append(region.rules.patterns, searchDef.rules.patterns...)
				region.rules.regions = append(region.rules.regions, searchDef.rules.regions...)
			}
		}
	}
	for _, r := range region.rules.regions {
		resolveIncludesInRegion(files, r)
		r.parent = region
	}
}

func parseRules(input []interface{}, curRegion *region) (ru *rules, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()
	ru = new(rules)

	for _, v := range input {
		rule := v.(map[interface{}]interface{})
		for k, val := range rule {
			group := k

			switch object := val.(type) {
			case string:
				if k == "include" {
					ru.includes = append(ru.includes, object)
				} else {
					// Pattern
					r, err := regexp.Compile(object)
					if err != nil {
						return nil, err
					}

					groupStr := group.(string)
					if _, ok := Groups[groupStr]; !ok {
						numGroups++
						Groups[groupStr] = numGroups
					}
					groupNum := Groups[groupStr]
					ru.patterns = append(ru.patterns, &pattern{groupNum, r})
				}
			case map[interface{}]interface{}:
				// region
				region, err := parseRegion(group.(string), object, curRegion)
				if err != nil {
					return nil, err
				}
				ru.regions = append(ru.regions, region)
			default:
				return nil, fmt.Errorf("Bad type %T", object)
			}
		}
	}

	return ru, nil
}

func parseRegion(group string, regionInfo map[interface{}]interface{}, prevRegion *region) (r *region, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()

	r = new(region)
	if _, ok := Groups[group]; !ok {
		numGroups++
		Groups[group] = numGroups
	}
	groupNum := Groups[group]
	r.group = groupNum
	r.parent = prevRegion

	r.start, err = regexp.Compile(regionInfo["start"].(string))

	if err != nil {
		return nil, err
	}

	r.end, err = regexp.Compile(regionInfo["end"].(string))

	if err != nil {
		return nil, err
	}

	// skip is optional
	if _, ok := regionInfo["skip"]; ok {
		r.skip, err = regexp.Compile(regionInfo["skip"].(string))

		if err != nil {
			return nil, err
		}
	}

	// limit-color is optional
	if _, ok := regionInfo["limit-group"]; ok {
		groupStr := regionInfo["limit-group"].(string)
		if _, ok := Groups[groupStr]; !ok {
			numGroups++
			Groups[groupStr] = numGroups
		}
		groupNum := Groups[groupStr]
		r.limitGroup = groupNum

		if err != nil {
			return nil, err
		}
	} else {
		r.limitGroup = r.group
	}

	r.rules, err = parseRules(regionInfo["rules"].([]interface{}), r)

	if err != nil {
		return nil, err
	}

	return r, nil
}
