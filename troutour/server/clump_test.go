package main

import (
	"golang.org/x/net/context"
	"testing"
	"time"
)

// A parallelogram of clumps.
// The potential-adjs nw-se and ne-sw intersect; the nw-se adj is shorter
// than ne-sw and should "win".
var fakeClumpDataParallelogram map[string]*Clump = map[string]*Clump{
	"sw": &Clump{"sw", 0, 110, 210, .011, .021, 0, []string{}, time.Now()},
	"nw": &Clump{"nw", 0, 120, 220, .012, .022, 0, []string{}, time.Now()},
	"se": &Clump{"se", 0, 110, 230, .011, .023, 0, []string{}, time.Now()},
	"ne": &Clump{"ne", 0, 120, 240, .012, .024, 0, []string{}, time.Now()},
}

func fakeQueryAdjFactory(x []ClumpAdj) func(ctx context.Context, query string) (cas []ClumpAdj, err error) {
	return func(ctx context.Context, query string) (cas []ClumpAdj, err error) {
		for _, ca := range x {
			if ca.EndIDs[0] == query || ca.EndIDs[1] == query {
				cas = append(cas, ca)
			}
		}
		return
	}
}

func TestACAREmptyParallelogram(t *testing.T) {
	clumps := fakeClumpDataParallelogram
	existingAdjs := []ClumpAdj{}
	queryAdjs := fakeQueryAdjFactory(existingAdjs)
	_, adds, rmIDs := adjifyComputeAddRms("sw", context.TODO(), clumps, queryAdjs)
	if len(rmIDs) > 0 {
		t.Error("wants to rm adjs from empty set")
	}
	if len(adds) != 3 {
		t.Error("should want to add 3 edges")
	}
}

func TestACARShouldBreakThrough(t *testing.T) {
	// potential se-nw adj is shorter than the existing sw-ne adj, and should "break" through
	clumps := fakeClumpDataParallelogram
	existingAdjs := []ClumpAdj{
		newAdj("nw", "sw"),
		newAdj("ne", "sw"),
		newAdj("se", "sw"),
	}
	queryAdjs := fakeQueryAdjFactory(existingAdjs)
	_, adds, rmIDs := adjifyComputeAddRms("se", context.TODO(), clumps, queryAdjs)
	if len(rmIDs) != 1 || rmIDs[0] != newAdj("ne", "sw").ID() {
		t.Error("should want to rm sw-ne adj")
	}
	if len(adds) != 2 {
		t.Error("should want to add 2 edges")
	}
}

func TestACARShouldBeBlocked(t *testing.T) {
	// potential sw-ne adj is longer than existing se-nw adj, should be blocked
	clumps := fakeClumpDataParallelogram
	existingAdjs := []ClumpAdj{
		newAdj("nw", "se"),
		newAdj("ne", "se"),
		newAdj("se", "sw"),
	}
	queryAdjs := fakeQueryAdjFactory(existingAdjs)
	_, adds, rmIDs := adjifyComputeAddRms("sw", context.TODO(), clumps, queryAdjs)
	// se-nw adj is shorter than the existing sw-ne adj, & should "break" through
	if len(rmIDs) > 0 {
		t.Error("shouldn't want to rm anything, but does")
	}
	if len(adds) != 1 || adds[0].ID() != newAdj("nw", "sw").ID() {
		t.Error("should want to add nw-sw adj (but not sw-ne, too long) got ", adds)
	}
}

func TestAFBEmpty(t *testing.T) {
	clumps := fakeClumpDataParallelogram
	existingAdjs := map[string]ClumpAdj{}
	adj := newAdj("sw", "ne")
	already, unblocked, blockerIDs := adjifyFindBlockers(adj, existingAdjs, clumps)
	if already {
		t.Error("Thinks new adj already loaded???")
	}
	if !unblocked {
		t.Error("Thinks blocked, but should be nothing there")
	}
	if len(blockerIDs) > 0 {
		t.Errorf("Thinks blocked by %v, but should be nothing there", blockerIDs)
	}
}

func TestAFBAlready(t *testing.T) {
	clumps := fakeClumpDataParallelogram
	existingAdj := newAdj("sw", "ne")
	existingAdjs := map[string]ClumpAdj{existingAdj.ID(): existingAdj}
	adj := newAdj("sw", "ne")
	already, _, _ := adjifyFindBlockers(adj, existingAdjs, clumps)
	if !already {
		t.Error("Doesn't notice already-loaded adj")
	}
}

func TestAFBShouldBeThwarted(t *testing.T) {
	clumps := fakeClumpDataParallelogram
	existingAdj := newAdj("se", "nw")
	existingAdjs := map[string]ClumpAdj{existingAdj.ID(): existingAdj}
	adj := newAdj("sw", "ne")
	already, unblocked, _ := adjifyFindBlockers(adj, existingAdjs, clumps)
	if already {
		t.Error("Thinks already loaded, but shouldn't be")
	}
	if unblocked {
		t.Error("Didn't notice blocker")
	}
}

func TestAFBShouldBreakThrough(t *testing.T) {
	clumps := fakeClumpDataParallelogram
	existingAdj := newAdj("sw", "ne")
	existingAdjs := map[string]ClumpAdj{existingAdj.ID(): existingAdj}
	adj := newAdj("se", "nw")
	already, unblocked, blockerIDs := adjifyFindBlockers(adj, existingAdjs, clumps)
	if already {
		t.Error("Thinks already loaded, but shouldn't be")
	}
	if !unblocked {
		t.Error("Should break through, but thinks blocked")
	}
	if len(blockerIDs) != 1 || blockerIDs[0] != existingAdj.ID() {
		t.Errorf("Got wrong blockers: %v", blockerIDs)
	}
}
