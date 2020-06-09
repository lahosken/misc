package main

import (
	"testing"
)

var tangle = []string{
	"alfa", "bravo", "charlie",
	"alfa", "delta", "charlie", "echo",
	"delta", "foxtrot", "charlie"}

func TestCompactAlfa(t *testing.T) {
	rm := routeListRemove(tangle, []string{"alfa"})
	rmGoal := []string{
		"", "bravo", "charlie",
		"", "delta", "charlie", "echo",
		"delta", "foxtrot", "charlie"}
	if len(rm) != len(rmGoal) {
		t.Error("rm lens don't match")
		return
	}
	for i, rmS := range rm {
		if rmS != rmGoal[i] {
			t.Errorf("rm item not match %d (%s) (%s).", i, rmS, rmGoal[i])
		}
	}
	compact := routeListCompact(rm)
	compactGoal := []string{
		"bravo", "charlie",
		"", "delta", "charlie", "echo",
		"delta", "foxtrot", "charlie"}
	if len(compact) != len(compactGoal) {
		t.Error("compact lens don't match")
		return
	}
	for i, compactS := range compact {
		if compactS != compactGoal[i] {
			t.Errorf("compact item not match %d (%s) (%s).", i, compactS, compactGoal[i])
		}
	}
}

func TestCompactBravo(t *testing.T) {
	rm := routeListRemove(tangle, []string{"bravo"})
	rmGoal := []string{
		"alfa", "", "charlie",
		"alfa", "delta", "charlie", "echo",
		"delta", "foxtrot", "charlie"}
	if len(rm) != len(rmGoal) {
		t.Error("rm lens don't match")
		return
	}
	for i, rmS := range rm {
		if rmS != rmGoal[i] {
			t.Errorf("rm item not match %d (%s) (%s).", i, rmS, rmGoal[i])
		}
	}
	compact := routeListCompact(rm)
	compactGoal := []string{
		"charlie",
		"alfa", "delta", "charlie", "echo",
		"delta", "foxtrot", "charlie"}
	if len(compact) != len(compactGoal) {
		t.Error("compact lens don't match")
		return
	}
	for i, compactS := range compact {
		if compactS != compactGoal[i] {
			t.Errorf("compact item not match %d (%s) (%s).", i, compactS, compactGoal[i])
		}
	}
}

func TestCompactCharlie(t *testing.T) {
	rm := routeListRemove(tangle, []string{"charlie"})
	rmGoal := []string{
		"alfa", "bravo", "",
		"alfa", "delta", "", "echo",
		"delta", "foxtrot", ""}
	if len(rm) != len(rmGoal) {
		t.Error("rm lens don't match")
		return
	}
	for i, rmS := range rm {
		if rmS != rmGoal[i] {
			t.Errorf("rm item not match %d (%s) (%s).", i, rmS, rmGoal[i])
		}
	}
	compact := routeListCompact(rm)
	compactGoal := []string{
		"alfa", "bravo", "",
		"alfa", "delta", "", "echo",
		"delta", "foxtrot"}
	if len(compact) != len(compactGoal) {
		t.Error("compact lens don't match")
		return
	}
	for i, compactS := range compact {
		if compactS != compactGoal[i] {
			t.Errorf("compact item not match %d (%s) (%s).", i, compactS, compactGoal[i])
		}
	}
}

func TestCompactFoxtrot(t *testing.T) {
	rm := routeListRemove(tangle, []string{"foxtrot"})
	rmGoal := []string{
		"alfa", "bravo", "charlie",
		"alfa", "delta", "charlie", "echo",
		"delta", "", "charlie"}
	if len(rm) != len(rmGoal) {
		t.Error("rm lens don't match")
		return
	}
	for i, rmS := range rm {
		if rmS != rmGoal[i] {
			t.Errorf("rm item not match %d (%s) (%s).", i, rmS, rmGoal[i])
		}
	}
	compact := routeListCompact(rm)
	compactGoal := []string{
		"alfa", "bravo", "charlie",
		"alfa", "delta", "charlie", "echo",
		"delta"}
	if len(compact) != len(compactGoal) {
		t.Error("compact lens don't match")
		return
	}
	for i, compactS := range compact {
		if compactS != compactGoal[i] {
			t.Errorf("compact item not match %d (%s) (%s).", i, compactS, compactGoal[i])
		}
	}
}

func TestCompactAlfaBravo(t *testing.T) {
	rm := routeListRemove(tangle, []string{"alfa", "bravo"})
	rmGoal := []string{
		"", "", "charlie",
		"", "delta", "charlie", "echo",
		"delta", "foxtrot", "charlie"}
	if len(rm) != len(rmGoal) {
		t.Error("rm lens don't match")
		return
	}
	for i, rmS := range rm {
		if rmS != rmGoal[i] {
			t.Errorf("rm item not match %d (%s) (%s).", i, rmS, rmGoal[i])
		}
	}
	compact := routeListCompact(rm)
	compactGoal := []string{
		// first 'row' gone
		"delta", "charlie", "echo",
		"delta", "foxtrot", "charlie"}
	if len(compact) != len(compactGoal) {
		t.Error("compact lens don't match")
		return
	}
	for i, compactS := range compact {
		if compactS != compactGoal[i] {
			t.Errorf("compact item not match %d (%s) (%s).", i, compactS, compactGoal[i])
		}
	}
}

func TestCompactAlfaCharlie(t *testing.T) {
	rm := routeListRemove(tangle, []string{"alfa", "charlie"})
	rmGoal := []string{
		"", "bravo", "",
		"", "delta", "", "echo",
		"delta", "foxtrot", ""}
	if len(rm) != len(rmGoal) {
		t.Error("rm lens don't match")
		return
	}
	for i, rmS := range rm {
		if rmS != rmGoal[i] {
			t.Errorf("rm item not match %d (%s) (%s).", i, rmS, rmGoal[i])
		}
	}
	compact := routeListCompact(rm)
	compactGoal := []string{
		"echo", "delta", "foxtrot"}
	if len(compact) != len(compactGoal) {
		t.Error("compact lens don't match")
		return
	}
	for i, compactS := range compact {
		if compactS != compactGoal[i] {
			t.Errorf("compact item not match %d (%s) (%s).", i, compactS, compactGoal[i])
		}
	}
}
