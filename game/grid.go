package game

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

func generateGrid(width, height, mineCount int) ([][]cell, error) {
	if width < 2 || height < 2 {
		return nil, fmt.Errorf("Invalid dimensions %dx%d. Must be at least 2x2.", width, height)
	}

	if width > 40 || height > 40 {
		return nil, fmt.Errorf("Invalid dimensions %dx%d. Must be at most 40x40.", width, height)
	}

	if mineCount < 1 {
		return nil, fmt.Errorf("Too few mintes (%d). Place at least 1.", mineCount)
	}

	if mineCount > width*height {
		return nil, fmt.Errorf("Too many mines (%d). The mine count cannot exceed the number of cells.", mineCount)
	}

	// Create a mine-less matrix all of zeroes.
	matrix := initEmptyGrid(width, height)

	// Decide on where to place mines.
	mineCoords := chooseMinePlacements(width, height, mineCount)
	for _, c := range mineCoords {
		matrix[c[0]][c[1]].isMined = true

		// Increment all adjacent cells' mine counts.
		if c[0] > 0 {
			if c[1] > 0 {
				matrix[c[0]-1][c[1]-1].adjacentMines++
			}
			matrix[c[0]-1][c[1]].adjacentMines++
			if c[1] < height-1 {
				matrix[c[0]-1][c[1]+1].adjacentMines++
			}
		}
		if c[1] > 0 {
			matrix[c[0]][c[1]-1].adjacentMines++
		}
		if c[1] < height-1 {
			matrix[c[0]][c[1]+1].adjacentMines++
		}
		if c[0] < width-1 {
			if c[1] > 0 {
				matrix[c[0]+1][c[1]-1].adjacentMines++
			}
			matrix[c[0]+1][c[1]].adjacentMines++
			if c[1] < height-1 {
				matrix[c[0]+1][c[1]+1].adjacentMines++
			}
		}
	}

	return matrix, nil
}

func initEmptyGrid(width, height int) [][]cell {
	matrix := make([][]cell, height)
	for i := 0; i < height; i++ {
		matrix[i] = make([]cell, width)
	}

	return matrix
}

func chooseMinePlacements(width, height, mineCount int) []coordinate {
	// Randomly choose row and column to place each mine.
	set := make(map[coordinate]bool)
	for ; mineCount > 0; mineCount-- {

		c := coordinate{
			int(rand.Float32() * float32(width)),
			int(rand.Float32() * float32(height)),
		}
		if set[c] == true {
			mineCount++
		} else {
			set[c] = true
		}
	}

	coords := make([]coordinate, 0, height)
	for k := range set {
		coords = append(coords, k)
	}

	return coords
}

// getNeighbors() will provide a list of all coordinates adjacent to the provided coordinate
// in a grid of the given dimensions.
func getNeighbors(coord coordinate, width, height int) []coordinate {
	neighbors := []coordinate{}
	xs := []int{coord[0]}
	ys := []int{coord[1]}

	if coord[0] > 0 {
		xs = append(xs, coord[0]-1)
	}
	if coord[0] < width-1 {
		xs = append(xs, coord[0]+1)
	}
	if coord[1] > 0 {
		ys = append(ys, coord[1]-1)
	}
	if coord[1] < height-1 {
		ys = append(ys, coord[1]+1)
	}

	for _, x := range xs {
		for _, y := range ys {
			c := coordinate{x, y}
			if c != coord {
				neighbors = append(neighbors, c)
			}
		}
	}

	return neighbors
}

func cellNameToCoordinate(cellName CellName) (coordinate, error) {
	// Must be letters followed by numbers.
	matches := validCellName.FindStringSubmatch(string(cellName))
	if matches == nil {
		return [2]int{0, 0}, fmt.Errorf("Invalid cell name '%s'. Must be a letter followed by a number, e.g., B6.", cellName)
	}

	// Convert letter to x
	x := columnKeyToInt(matches[1])

	// Convert number to y
	y, err := strconv.Atoi(matches[2])
	if err != nil {
		return [2]int{0, 0}, fmt.Errorf("Invalid cell name '%s': %s", cellName, err)
	}
	y--

	return [2]int{x, y}, nil
}

// columnKeyToInt() converts a column key (e.g., AA) to an integer starting at 0.
func columnKeyToInt(columnKey string) int {
	// Uppercase only so that we can subtract exactly 64 from the ASCII code.
	columnKey = strings.ToUpper(columnKey)

	x := 0
	place := len(columnKey) - 1
	for _, char := range columnKey {
		x += (int(char) - 64) * int(math.Pow(26, float64(place)))
		place--
	}

	// Subtract 1 so that A = 0.
	return x - 1
}

func containsCoordinate(coord coordinate, grid [][]cell) bool {
	return coord[0] >= 0 &&
		coord[0] < len(grid[0]) &&
		coord[1] >= 0 &&
		coord[1] < len(grid)
}
