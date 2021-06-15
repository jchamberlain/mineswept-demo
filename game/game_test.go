package game

import (
	"sort"
	"testing"
)

func TestNewGameShouldErrorOnTooWide(t *testing.T) {
	_, err := NewGame(41, 20, 10)
	if err == nil {
		t.Error("Expected error if width greater than 40.")
	}
}

func TestNewGameShouldErrorOnTooNarrow(t *testing.T) {
	_, err := NewGame(1, 20, 10)
	if err == nil {
		t.Error("Expected error if width less than 2.")
	}
}

func TestNewGameShouldErrorOnTooTall(t *testing.T) {
	_, err := NewGame(20, 41, 10)
	if err == nil {
		t.Error("Expected error if height greater than 40.")
	}
}

func TestNewGameShouldErrorOnTooShort(t *testing.T) {
	_, err := NewGame(20, 1, 10)
	if err == nil {
		t.Error("Expected error if height less than 2.")
	}
}

func TestNewGameShouldErrorOnTooManyMines(t *testing.T) {
	_, err := NewGame(20, 20, 401)
	if err == nil {
		t.Error("Expected error for more mines than cells.")
	}
}

func TestNewGame(t *testing.T) {
	g, err := NewGame(10, 10, 10)
	if err != nil {
		t.Errorf("Unexpected error generating game: %s", err)
	}

	// Check that an ID is populated with version 1.
	if len(g.id) != 36 {
		t.Errorf("Invalid game ID: %s", g.id)
	}
	if g.version != 1 {
		t.Errorf("Game version should start at 1 (is %d)", g.version)
	}

	// Check that the grid matches the desired size and number of mines.
	if len(g.grid) != 10 {
		t.Errorf("Game grid should have height of 10 (is %d)", len(g.grid))
	}
	if len(g.grid[0]) != 10 {
		t.Errorf("Game grid should have width of 10 (is %d)", len(g.grid[0]))
	}

	mineCount := 0
	revealedCount := 0
	for _, row := range g.grid {
		for _, cell := range row {
			if cell.isMined {
				mineCount += 1
			}
			if cell.isRevealed {
				revealedCount += 1
			}
		}
	}
	if mineCount != 10 {
		t.Errorf("Game grid has incorrect number of mines (expected 10, found %d)", mineCount)
	}
	if revealedCount != 0 {
		t.Errorf("Game grid should start with no cells revealed (found %d revealed", revealedCount)
	}

	// Check that one event is added to the log.
	if len(g.events) != 1 {
		t.Errorf("Game should start with one event in log (found %d)", len(g.events))
	}

	switch v := g.events[0].(type) {
	case gameStartedEvent:
		// Good!
	default:
		t.Errorf("Game's first event should be a gameStartedEvent (is %T)", v)
	}
}

func TestRevealCell(t *testing.T) {
	// Create a new game, but then hijack and re-apply its first event to set our own predetermined grid.
	g, _ := NewGame(5, 5, 5)
	event := g.events[0].(gameStartedEvent)
	event.grid = makeExampleGrid()
	g.events[0] = event
	event.applyTo(g)

	// Try a non-existent cell.
	err := g.RevealCell("Z30")
	if err == nil {
		t.Errorf("Failed to detect non-existent cell")
	}

	// Try a non-mined cell.
	err = g.RevealCell("A1")
	if err != nil {
		t.Errorf("Failed to reveal cell A1: %s", err)
	}

	if g.grid[0][0].isRevealed != true {
		t.Error("Failed to reveal cell A1")
	}

	if g.grid[0][0].isFlagged != false {
		t.Error("Incorrectly flagged cell A1")
	}

	// Try a non-mined cell with no adjacent mines. (Should reveal neighboring cells recursively.)
	err = g.RevealCell("E3")
	if err != nil {
		t.Errorf("Failed to reveal cell E3: %s", err)
	}

	expectedRevealed := []CellName{"D2", "E2", "D3", "D4", "E4", "C2", "C3", "C4"}
	for _, cellName := range expectedRevealed {
		coord, _ := cellNameToCoordinate(cellName)
		if !g.grid[coord[1]][coord[0]].isRevealed {
			t.Errorf("Failed to reveal neighboring cell %s (%d,%d)", cellName, coord[0], coord[1])
		}
	}

	// Try a previously revealed cell.
	g.grid[0][0].isRevealed = true
	err = g.RevealCell("A1")
	if err == nil {
		t.Errorf("Failed to detect previously revealed cell")
	}

	// Try a mined cell.
	err = g.RevealCell("B2")
	if err != nil {
		t.Errorf("Failed to reveal cell B2: %s", err)
	}

	if g.grid[1][1].isRevealed != true {
		t.Error("Failed to reveal cell B2")
	}

	if g.grid[1][1].isFlagged != false {
		t.Error("Incorrectly flagged cell B2")
	}

	// Check that ALL cells are revealed.
	for i := 0; i < len(g.grid); i++ {
		for j := 0; j < len(g.grid[i]); j++ {
			if g.grid[i][j].isRevealed != true {
				t.Errorf("All cells should be revealed: failed to reveal %d,%d", j, i)
			}
		}
	}
}

func TestGetNeighbors(t *testing.T) {
	// Top-left corner
	neighbors := getNeighbors(coordinate{0, 0}, 5, 5)
	expected := []coordinate{
		coordinate{1, 0},
		coordinate{0, 1},
		coordinate{1, 1},
	}
	assertEqualCoords("Should get neighbors for top-left cell", expected, neighbors, t)

	// Top-right corner
	neighbors = getNeighbors(coordinate{4, 0}, 5, 5)
	expected = []coordinate{
		coordinate{3, 0},
		coordinate{3, 1},
		coordinate{4, 1},
	}
	assertEqualCoords("Should get neighbors for top-right cell", expected, neighbors, t)

	// Bottom-left corner
	neighbors = getNeighbors(coordinate{0, 4}, 5, 5)
	expected = []coordinate{
		coordinate{0, 3},
		coordinate{1, 3},
		coordinate{1, 4},
	}
	assertEqualCoords("Should get neighbors for bottom-left cell", expected, neighbors, t)

	// Bottom-right corner
	neighbors = getNeighbors(coordinate{4, 4}, 5, 5)
	expected = []coordinate{
		coordinate{3, 3},
		coordinate{4, 3},
		coordinate{3, 4},
	}
	assertEqualCoords("Should get neighbors for bottom-right cell", expected, neighbors, t)

	// A left side
	neighbors = getNeighbors(coordinate{0, 2}, 5, 5)
	expected = []coordinate{
		coordinate{0, 1},
		coordinate{1, 1},
		coordinate{1, 2},
		coordinate{0, 3},
		coordinate{1, 3},
	}
	assertEqualCoords("Should get neighbors for a left side cell", expected, neighbors, t)

	// Somewhere in the middle
	neighbors = getNeighbors(coordinate{2, 2}, 5, 5)
	expected = []coordinate{
		coordinate{1, 1},
		coordinate{2, 1},
		coordinate{3, 1},
		coordinate{1, 2},
		coordinate{3, 2},
		coordinate{1, 3},
		coordinate{2, 3},
		coordinate{3, 3},
	}
	assertEqualCoords("Should get neighbors for an inner cell", expected, neighbors, t)
}

func TestColumnKeyToInt(t *testing.T) {
	i := columnKeyToInt("A")
	if i != 0 {
		t.Errorf("Expected 0 for A, got %d", i)
	}

	i = columnKeyToInt("a")
	if i != 0 {
		t.Errorf("Expected 0 for a, got %d", i)
	}

	i = columnKeyToInt("B")
	if i != 1 {
		t.Errorf("Expected 1 for B, got %d", i)
	}

	i = columnKeyToInt("Z")
	if i != 25 {
		t.Errorf("Expected 25 for Z, got %d", i)
	}

	i = columnKeyToInt("AA")
	if i != 26 {
		t.Errorf("Expected 26 for AA, got %d", i)
	}

	i = columnKeyToInt("BZ")
	if i != 77 {
		t.Errorf("Expected 77 for BZ, got %d", i)
	}
}

func TestCellNameToCoord(t *testing.T) {
	coord, err := cellNameToCoordinate("A1")
	if err != nil {
		t.Errorf("Failed converting cell name A1: %s", err)
	} else if coord[0] != 0 || coord[1] != 0 {
		t.Errorf("Expected 0,0 for cell name A1, got %d,%d", coord[0], coord[1])
	}

	coord, err = cellNameToCoordinate("b2")
	if err != nil {
		t.Errorf("Failed converting cell name b2: %s", err)
	} else if coord[0] != 1 || coord[1] != 1 {
		t.Errorf("Expected 1,1 for cell name b2, got %d,%d", coord[0], coord[1])
	}
}

func makeExampleGrid() [][]cell {
	// 1  1  2  X  1
	// 1  X  2  1  1
	// 3  3  2  0  0
	// X  X  1  1  1
	// 2  2  1  1  X

	return [][]cell{
		[]cell{
			cell{isMined: false, adjacentMines: 1},
			cell{isMined: false, adjacentMines: 1},
			cell{isMined: false, adjacentMines: 2},
			cell{isMined: true},
			cell{isMined: false, adjacentMines: 1},
		},
		[]cell{
			cell{isMined: false, adjacentMines: 1},
			cell{isMined: true},
			cell{isMined: false, adjacentMines: 2},
			cell{isMined: false, adjacentMines: 1},
			cell{isMined: false, adjacentMines: 1},
		},
		[]cell{
			cell{isMined: false, adjacentMines: 3},
			cell{isMined: false, adjacentMines: 3},
			cell{isMined: false, adjacentMines: 2},
			cell{isMined: false, adjacentMines: 0},
			cell{isMined: false, adjacentMines: 0},
		},
		[]cell{
			cell{isMined: true},
			cell{isMined: true},
			cell{isMined: false, adjacentMines: 1},
			cell{isMined: false, adjacentMines: 1},
			cell{isMined: false, adjacentMines: 1},
		},
		[]cell{
			cell{isMined: false, adjacentMines: 2},
			cell{isMined: false, adjacentMines: 2},
			cell{isMined: false, adjacentMines: 1},
			cell{isMined: false, adjacentMines: 1},
			cell{isMined: true},
		},
	}
}

func assertEqualCoords(msg string, expected, found []coordinate, t *testing.T) {
	if len(expected) != len(found) {
		t.Errorf("%s\nExpected %s\nFound    %s", msg, expected, found)
		return
	}

	expectedStrings := make([]string, len(expected))
	for _, coord := range expected {
		expectedStrings = append(expectedStrings, coord.String())
	}
	sort.Strings(expectedStrings)

	foundStrings := make([]string, len(found))
	for _, coord := range found {
		foundStrings = append(foundStrings, coord.String())
	}
	sort.Strings(foundStrings)

	for i := 0; i < len(expectedStrings); i++ {
		if expectedStrings[i] != foundStrings[i] {
			t.Errorf("%s\nExpected %s\nFound    %s", msg, expected, found)
			return
		}
	}
}
