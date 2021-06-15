package game

import (
	"fmt"
	"regexp"
	"time"

	"zephyri.co/mineswept/eventsource"
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
	e := gameStartedEvent{
		BaseEvent: eventsource.BaseEvent{
			AggregateId: eventsource.NewAggregateId(),
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

	if !containsCoordinate(coord, g.grid) {
		return fmt.Errorf("Invalid cell %s (%d,%d).", cellName, coord[0], coord[1])
	}

	if g.grid[coord[1]][coord[0]].isRevealed {
		return fmt.Errorf("Cell %s already revealed", cellName)
	}

	// Generate and apply a simple cell reveal event.
	revealed := cellRevealedEvent{
		BaseEvent: eventsource.BaseEvent{
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
		BaseEvent: eventsource.BaseEvent{
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
		BaseEvent: eventsource.BaseEvent{
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
				BaseEvent: eventsource.BaseEvent{
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

type cell struct {
	isFlagged     bool
	isMined       bool
	isRevealed    bool
	adjacentMines int
}

type event interface {
	applyTo(g *game)
}

type gameStartedEvent struct {
	eventsource.BaseEvent
	grid [][]cell
}

func (e gameStartedEvent) applyTo(g *game) {
	g.onGameStarted(e)
}

type cellRevealedEvent struct {
	eventsource.BaseEvent
	InteractionCellName CellName
	CellCoord           coordinate
}

func (e cellRevealedEvent) applyTo(g *game) {
	g.onCellRevealed(e)
}

type cellFlaggedEvent struct{}

type gameWonEvent struct {
	eventsource.BaseEvent
}

func (e gameWonEvent) applyTo(g *game) {
	g.onGameWon(e)
}

type gameLostEvent struct {
	eventsource.BaseEvent
}

func (e gameLostEvent) applyTo(g *game) {
	g.onGameLost(e)
}
