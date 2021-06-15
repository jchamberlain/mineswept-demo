package game

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var validCellName = regexp.MustCompile("([A-z]+)([0-9]+)")

type Game interface {
	IsComplete() bool
}

// type AggregateId string

type GameInfo struct {
	Id   string
	Name string
}

// ListSavedGames will look in a hidden directory in the user's home for any previously saved games.
func ListSavedGames() []GameInfo {
	return []GameInfo{{"asdf", "First Game"}}
}

type game struct {
	id                         string
	version                    int
	name                       string
	grid                       [][]cell
	cellCount                  int
	revealedOrFlaggedCellCount int
	isEnded                    bool
	createAt                   time.Time
	updatedAt                  time.Time
	events                     []event
}

type CellName string

type coordinate [2]int

func (c coordinate) String() string {
	return fmt.Sprintf("%d,%d", c[0], c[1])
}

// NewGame will create a new game with a grid initialized to the desired size and mine count.
func NewGame(width, height, mineCount int) (*game, error) {
	// Initialize a valid grid if possible, else return an error.
	grid, err := generateGrid(width, height, mineCount)
	if err != nil {
		return nil, err
	}

	// Make the initial Game model.
	g := game{}

	// Append the first event with the complete initial state.
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("Unable to generate a UUID! %s", err)
	}

	e := gameStartedEvent{
		BaseEvent: BaseEvent{
			AggregateId: id.String(),
			Version:     1,
			At:          time.Now(),
		},
		grid: grid,
	}
	e.applyTo(&g)
	g.events = append(g.events, e)

	return &g, nil
}

func (g *game) onGameStarted(e gameStartedEvent) []event {
	g.id = e.AggregateId
	g.version = e.Version
	g.grid = e.grid
	g.cellCount = len(g.grid) * len(g.grid[0])
	return []event{}
}

// RevealCell makes a cell visible. If it's mined, you blow up!
func (g *game) RevealCell(cellName CellName) error {
	// Check that this is a valid move before generating an event.
	coord, err := cellNameToCoordinate(cellName)
	if err != nil {
		return err
	}

	if !g.containsCoordinate(coord) {
		return fmt.Errorf("Invalid cell %s (%d,%d).", cellName, coord[0], coord[1])
	}

	if g.grid[coord[1]][coord[0]].isRevealed {
		return fmt.Errorf("Cell %s already revealed", cellName)
	}

	// Generate and apply a simple cell reveal event.
	revealed := cellRevealedEvent{
		BaseEvent: BaseEvent{
			AggregateId: g.id,
			Version:     g.version + 1,
			At:          time.Now(),
		},
		InteractionCellName: cellName,
		CellCoord:           coord,
	}
	revealed.applyTo(g)

	// With that cell now revealed, generate and apply additional events if we've stepped
	// on a mine (lost), correctly played the last cell (won), or need to automatically
	// reveal additional cells.
	//
	// Each called method will generate and apply the events themselves, returning them so
	// we can persist events as desired.
	if lost := g.loseGameIfMined(coord); lost != nil {
		g.events = append(g.events, lost)
		return nil
	}

	if revealedNeighbors := g.revealNeighborsIfNoAdjacentMines(coord, revealed); len(revealedNeighbors) > 0 {
		g.events = append(g.events, revealedNeighbors...)
	}

	if won := g.winGameIfLastCell(coord); won != nil {
		g.events = append(g.events, won)

	}

	return nil
}

func (g *game) onCellRevealed(e cellRevealedEvent) {
	target := &g.grid[e.CellCoord[1]][e.CellCoord[0]]
	target.isRevealed = true
	g.revealedOrFlaggedCellCount++
}

func (g *game) onGameLost(e gameLostEvent) {
	// Mark game as ended and reveal all cells.
	g.isEnded = true

	for y := 0; y < len(g.grid); y++ {
		for x := 0; x < len(g.grid[y]); x++ {
			if !g.grid[y][x].isRevealed && !g.grid[y][x].isFlagged {
				g.revealedOrFlaggedCellCount++
			}

			g.grid[y][x].isRevealed = true
		}
	}
}

func (g *game) onGameWon(e gameWonEvent) {
	// Mark game as ended.
	g.isEnded = true
}

func (g *game) loseGameIfMined(coord coordinate) event {
	target := g.grid[coord[1]][coord[0]]

	if !target.isMined || !target.isRevealed {
		return nil
	}

	e := gameLostEvent{
		BaseEvent: BaseEvent{
			AggregateId: g.id,
			Version:     g.version + 1,
			At:          time.Now(),
		},
	}
	e.applyTo(g)

	return e
}

func (g *game) winGameIfLastCell(coord coordinate) event {
	if g.cellCount != g.revealedOrFlaggedCellCount {
		return nil
	}

	e := gameWonEvent{
		BaseEvent: BaseEvent{
			AggregateId: g.id,
			Version:     g.version + 1,
			At:          time.Now(),
		},
	}
	e.applyTo(g)

	return e
}

func (g *game) revealNeighborsIfNoAdjacentMines(coord coordinate, originalEvent cellRevealedEvent) []event {
	events := []event{}

	// If there are adjacent mines, do nothing.
	if g.grid[coord[1]][coord[0]].adjacentMines > 0 {
		return events
	}

	// If there are no adjacent mines, reveal neighboring cells. Repeat for any
	// neighbor with no adjacent mines (breadth-first traversal of the graph).
	//
	// For each new cell which needs to be revealed, apply and emit an event.
	queue := getNeighbors(coord, len(g.grid[0]), len(g.grid))
	for i := 0; i < len(queue); i++ {
		neighbor := &g.grid[queue[i][1]][queue[i][0]]

		if !neighbor.isRevealed && !neighbor.isMined {
			revealed := cellRevealedEvent{
				BaseEvent: BaseEvent{
					AggregateId: g.id,
					Version:     g.version + 1,
					At:          time.Now(),
				},
				InteractionCellName: originalEvent.InteractionCellName,
				CellCoord:           queue[i],
			}
			revealed.applyTo(g)
			events = append(events, revealed)

			// If this newly revealed cell also has no adjacent mines, keep going!
			if neighbor.adjacentMines == 0 {
				queue = append(queue, getNeighbors(queue[i], len(g.grid[0]), len(g.grid))...)
			}
		}
	}

	return events
}

func (g *game) FlagCell() {}

func (g *game) UndoMove() {}

func (g *game) IsComplete() bool {
	return false
}

func (g *game) containsCoordinate(coord coordinate) bool {
	return coord[0] >= 0 &&
		coord[0] < len(g.grid[0]) &&
		coord[1] >= 0 &&
		coord[1] < len(g.grid)
}

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
	matrix := initEmptyMatrix(width, height)

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

func initEmptyMatrix(width, height int) [][]cell {
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

type cell struct {
	isFlagged     bool
	isMined       bool
	isRevealed    bool
	adjacentMines int
}

type event interface {
	applyTo(g *game)
}

type BaseEvent struct {
	AggregateId string
	Version     int
	At          time.Time
}

type gameStartedEvent struct {
	BaseEvent
	grid [][]cell
}

func (e gameStartedEvent) applyTo(g *game) {
	g.onGameStarted(e)
}

type cellRevealedEvent struct {
	BaseEvent
	InteractionCellName CellName
	CellCoord           coordinate
}

func (e cellRevealedEvent) applyTo(g *game) {
	g.onCellRevealed(e)
}

type cellFlaggedEvent struct{}

type gameWonEvent struct {
	BaseEvent
}

func (e gameWonEvent) applyTo(g *game) {

}

type gameLostEvent struct {
	BaseEvent
}

func (e gameLostEvent) applyTo(g *game) {
	g.onGameLost(e)
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
