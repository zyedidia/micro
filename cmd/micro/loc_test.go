package main

import (
	"testing"
)

func SquareTestPoints(t *testing.T, p1, p2 Loc) {
	// Test corners
	if !(Loc{10, 10}).Inside(p1, p2) {
		t.Error("Top-left corner should be inside")
	}
	if !(Loc{20, 20}).Inside(p1, p2) {
		t.Error("Bottom-right corner should be inside")
	}
	if !(Loc{10, 20}).Inside(p1, p2) {
		t.Error("Bottom-right corner should be inside")
	}
	if !(Loc{20, 10}).Inside(p1, p2) {
		t.Error("Top-right corner should be inside")
	}
	if !(Loc{15, 15}).Inside(p1, p2) {
		t.Error("Middle point should be inside")
	}
	if (Loc{9, 10}).Inside(p1, p2) {
		t.Error("Outside point 1 failed")
	}
	if (Loc{10, 9}).Inside(p1, p2) {
		t.Error("Outside point 2 failed")
	}
	if (Loc{20, 9}).Inside(p1, p2) {
		t.Error("Outside point 3 failed")
	}
	if (Loc{21, 10}).Inside(p1, p2) {
		t.Error("Outside point 4 failed")
	}
	if (Loc{21, 20}).Inside(p1, p2) {
		t.Error("Outside point 5 failed")
	}
	if (Loc{20, 21}).Inside(p1, p2) {
		t.Error("Outside point 6 failed")
	}
	if (Loc{10, 21}).Inside(p1, p2) {
		t.Error("Outside point 7 failed")
	}
	if (Loc{9, 20}).Inside(p1, p2) {
		t.Error("Outside point 8 failed")
	}
}

func TestLocInside(t *testing.T) {
	topLeft := Loc{10, 10}
	topRight := Loc{20, 10}
	bottomRight := Loc{20, 20}
	bottomLeft := Loc{10, 20}
	// Square made with top-left & bottom-right
	SquareTestPoints(t, topLeft, bottomRight)
	// Square made with top-right & bottom-left
	SquareTestPoints(t, topRight, bottomLeft)
	// Square made with bottom-right & top-left
	SquareTestPoints(t, bottomRight, topLeft)
	// Square made with bottom-left & top-right
	SquareTestPoints(t, bottomLeft, topRight)
}
