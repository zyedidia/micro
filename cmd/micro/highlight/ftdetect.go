package highlight

func DetectFiletype(defs []*Def, filename string, firstLine []byte) *Def {
	for _, d := range defs {
		if d.ftdetect[0].MatchString(filename) {
			return d
		}
		if len(d.ftdetect) > 1 {
			if d.ftdetect[1].MatchString(string(firstLine)) {
				return d
			}
		}
	}

	emptyDef := new(Def)
	emptyDef.FileType = "Unknown"
	emptyDef.rules = new(Rules)
	return emptyDef
}
