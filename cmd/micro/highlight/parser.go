package highlight

import (
	"fmt"
	"regexp"

	"gopkg.in/yaml.v2"
)

// A Def is a full syntax definition for a language
// It has a filetype, information about how to detect the filetype based
// on filename or header (the first line of the file)
// Then it has the rules which define how to highlight the file
type Def struct {
	FileType string
	ftdetect []*regexp.Regexp
	rules    *Rules
}

// A Pattern is one simple syntax rule
// It has a group that the rule belongs to, as well as
// the regular expression to match the pattern
type Pattern struct {
	group string
	regex *regexp.Regexp
}

// Rules defines which patterns and regions can be used to highlight
// a filetype
type Rules struct {
	regions  []*Region
	patterns []*Pattern
	includes []string
}

// A Region is a highlighted region (such as a multiline comment, or a string)
// It belongs to a group, and has start and end regular expressions
// A Region also has rules of its own that only apply when matching inside the
// region and also rules from the above region do not match inside this region
// Note that a region may contain more regions
type Region struct {
	group  string
	parent *Region
	start  *regexp.Regexp
	end    *regexp.Regexp
	rules  *Rules
}

// ParseDef parses an input syntax file into a highlight Def
func ParseDef(input []byte) (s *Def, err error) {
	// This is just so if we have an error, we can exit cleanly and return the parse error to the user
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	var rules map[interface{}]interface{}
	if err = yaml.Unmarshal(input, &rules); err != nil {
		return nil, err
	}

	s = new(Def)

	for k, v := range rules {
		if k == "filetype" {
			filetype := v.(string)

			s.FileType = filetype
		} else if k == "detect" {
			ftdetect := v.(map[interface{}]interface{})
			if len(ftdetect) >= 1 {
				syntax, err := regexp.Compile(ftdetect["filename"].(string))
				if err != nil {
					return nil, err
				}

				s.ftdetect = append(s.ftdetect, syntax)
			}
			if len(ftdetect) >= 2 {
				header, err := regexp.Compile(ftdetect["header"].(string))
				if err != nil {
					return nil, err
				}

				s.ftdetect = append(s.ftdetect, header)
			}
		} else if k == "rules" {
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

func ResolveIncludes(defs []*Def) {
	for _, d := range defs {
		resolveIncludesInDef(defs, d)
	}
}

func resolveIncludesInDef(defs []*Def, d *Def) {
	for _, lang := range d.rules.includes {
		for _, searchDef := range defs {
			if lang == searchDef.FileType {
				d.rules.patterns = append(d.rules.patterns, searchDef.rules.patterns...)
				d.rules.regions = append(d.rules.regions, searchDef.rules.regions...)
			}
		}
	}
	for _, r := range d.rules.regions {
		resolveIncludesInRegion(defs, r)
		r.parent = nil
	}
}

func resolveIncludesInRegion(defs []*Def, region *Region) {
	for _, lang := range region.rules.includes {
		for _, searchDef := range defs {
			if lang == searchDef.FileType {
				region.rules.patterns = append(region.rules.patterns, searchDef.rules.patterns...)
				region.rules.regions = append(region.rules.regions, searchDef.rules.regions...)
			}
		}
	}
	for _, r := range region.rules.regions {
		resolveIncludesInRegion(defs, r)
		r.parent = region
	}
}

func parseRules(input []interface{}, curRegion *Region) (*Rules, error) {
	rules := new(Rules)

	for _, v := range input {
		rule := v.(map[interface{}]interface{})
		for k, val := range rule {
			group := k

			switch object := val.(type) {
			case string:
				if k == "include" {
					rules.includes = append(rules.includes, object)
				} else {
					// Pattern
					r, err := regexp.Compile(object)
					if err != nil {
						return nil, err
					}

					rules.patterns = append(rules.patterns, &Pattern{group.(string), r})
				}
			case map[interface{}]interface{}:
				// Region
				region, err := parseRegion(group.(string), object, curRegion)
				if err != nil {
					return nil, err
				}
				rules.regions = append(rules.regions, region)
			default:
				return nil, fmt.Errorf("Bad type %T", object)
			}
		}
	}

	return rules, nil
}

func parseRegion(group string, regionInfo map[interface{}]interface{}, prevRegion *Region) (*Region, error) {
	var err error

	region := new(Region)
	region.group = group
	region.parent = prevRegion

	region.start, err = regexp.Compile(regionInfo["start"].(string))

	if err != nil {
		return nil, err
	}

	region.end, err = regexp.Compile(regionInfo["end"].(string))

	if err != nil {
		return nil, err
	}

	region.rules, err = parseRules(regionInfo["rules"].([]interface{}), region)

	if err != nil {
		return nil, err
	}

	return region, nil
}
