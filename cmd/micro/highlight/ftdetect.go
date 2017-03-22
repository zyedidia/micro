package highlight

func DetectFiletype(defs []*Def, filename string, firstLine []byte) *Def {
	for _, d := range defs {
		if isMatch, _ := d.ftdetect[0].MatchString(filename); isMatch {
			return d
		}
		if len(d.ftdetect) > 1 {
			if isMatch, _ := d.ftdetect[1].MatchString(string(firstLine)); isMatch {
				return d
			}
		}
	}

	emptyDef := new(Def)
	emptyDef.FileType = "Unknown"
	emptyDef.rules = new(Rules)
	return emptyDef
}
