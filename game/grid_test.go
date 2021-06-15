package game

import (
  "testing"
)

func TestGetNeighbors(t *testing.T) {
  // Top-left corner
  neighbors := getNeighbors(coordinate{0, 0}, 5, 5)
  expected := []coordinate{
    {1, 0},
    {0, 1},
    {1, 1},
  }
  assertEqualCoords("Should get neighbors for top-left cell", expected, neighbors, t)

  // Top-right corner
  neighbors = getNeighbors(coordinate{4, 0}, 5, 5)
  expected = []coordinate{
    {3, 0},
    {3, 1},
    {4, 1},
  }
  assertEqualCoords("Should get neighbors for top-right cell", expected, neighbors, t)

  // Bottom-left corner
  neighbors = getNeighbors(coordinate{0, 4}, 5, 5)
  expected = []coordinate{
    {0, 3},
    {1, 3},
    {1, 4},
  }
  assertEqualCoords("Should get neighbors for bottom-left cell", expected, neighbors, t)

  // Bottom-right corner
  neighbors = getNeighbors(coordinate{4, 4}, 5, 5)
  expected = []coordinate{
    {3, 3},
    {4, 3},
    {3, 4},
  }
  assertEqualCoords("Should get neighbors for bottom-right cell", expected, neighbors, t)

  // A left side
  neighbors = getNeighbors(coordinate{0, 2}, 5, 5)
  expected = []coordinate{
    {0, 1},
    {1, 1},
    {1, 2},
    {0, 3},
    {1, 3},
  }
  assertEqualCoords("Should get neighbors for a left side cell", expected, neighbors, t)

  // Somewhere in the middle
  neighbors = getNeighbors(coordinate{2, 2}, 5, 5)
  expected = []coordinate{
    {1, 1},
    {2, 1},
    {3, 1},
    {1, 2},
    {3, 2},
    {1, 3},
    {2, 3},
    {3, 3},
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
