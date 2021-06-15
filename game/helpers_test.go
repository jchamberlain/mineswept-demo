package game

import (
	"sort"
	"testing"
)

func makeExampleGrid() [][]cell {
	// 1  1  2  X  1
	// 1  X  2  1  1
	// 3  3  2  0  0
	// X  X  1  1  1
	// 2  2  1  1  X

	return [][]cell{
		{
			{isMined: false, adjacentMines: 1},
			{isMined: false, adjacentMines: 1},
			{isMined: false, adjacentMines: 2},
			{isMined: true},
			{isMined: false, adjacentMines: 1},
		},
		{
			{isMined: false, adjacentMines: 1},
			{isMined: true},
			{isMined: false, adjacentMines: 2},
			{isMined: false, adjacentMines: 1},
			{isMined: false, adjacentMines: 1},
		},
		{
			{isMined: false, adjacentMines: 3},
			{isMined: false, adjacentMines: 3},
			{isMined: false, adjacentMines: 2},
			{isMined: false, adjacentMines: 0},
			{isMined: false, adjacentMines: 0},
		},
		{
			{isMined: true},
			{isMined: true},
			{isMined: false, adjacentMines: 1},
			{isMined: false, adjacentMines: 1},
			{isMined: false, adjacentMines: 1},
		},
		{
			{isMined: false, adjacentMines: 2},
			{isMined: false, adjacentMines: 2},
			{isMined: false, adjacentMines: 1},
			{isMined: false, adjacentMines: 1},
			{isMined: true},
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
