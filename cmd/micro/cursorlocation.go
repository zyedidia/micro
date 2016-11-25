package main

type CursorLocation struct {
	X    int
	Y    int
	Path string
}

type CursorLocations struct {
	CursorLocations []CursorLocation
	Position        int
}

func (cl *CursorLocations) AddLocation(cursorLocation CursorLocation) {
	if len(cl.CursorLocations) > 500 {
		cl.CursorLocations = cl.CursorLocations[1:]
		cl.Position--
	}
	if len(cl.CursorLocations) < cl.Position-1 {
		cl.Position = len(cl.CursorLocations) - 1
	}
	cl.CursorLocations = cl.CursorLocations[:cl.Position]
	cl.CursorLocations = append(cl.CursorLocations, cursorLocation)
	cl.Position++
}

func (cl *CursorLocations) GetPrev() CursorLocation {
	if cl.Position > 0 {
		cl.Position--
	}

	if cl.Position >= len(cl.CursorLocations) {
		cl.Position = len(cl.CursorLocations) - 1
	}

	return cl.CursorLocations[cl.Position]
}

func (cl *CursorLocations) GetNext() CursorLocation {
	cl.Position++

	if cl.Position >= len(cl.CursorLocations) {
		cl.Position = len(cl.CursorLocations) - 1
	}
	return cl.CursorLocations[cl.Position]
}
