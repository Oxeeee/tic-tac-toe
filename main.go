package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
)

// Define constants for player symbols and game states
const xSymbol = "\x1b[34mX\x1b[0m"
const oSymbol = "\x1b[36mO\x1b[0m"
const noWinner = "none"

// Define winning combinations for the tic-tac-toe board
var winCombinations = [][]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // verticals
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // horizontals
	{0, 4, 8}, {2, 4, 6}, // diagonals
}

// Player struct represents a player in the game
type Player struct {
	Index      int
	Connection net.Conn
	Symbol     string
	Score      int
}

// Game struct represents the state of the game
type Game struct {
	Board         map[int]string
	Players       []Player
	CurrentPlayer Player
	Watchers      []net.Conn
}

// String method for Game to display the board and scores
func (g *Game) String() string {
	if !g.isFullPlaces() {
		return "waiting for an opponent to join \n"
	}
	var board, score string
	for i := 0; i < 9; i += 3 {
		board += fmt.Sprintf("%v | %v | %v\n", g.Board[i], g.Board[i+1], g.Board[i+2])
	}
	for _, p := range g.Players {
		score += fmt.Sprintf("%v:%v ", p.Symbol, p.Score)
	}
	return fmt.Sprintf("%v \nscore: %v\n", board, score)
}

// Checks if the board is full
func (g *Game) isFullBoard() bool {
	for _, v := range g.Board {
		if v != xSymbol && v != oSymbol {
			return false
		}
	}
	return true
}

// Checks if both players have joined the game
func (g *Game) isFullPlaces() bool {
	return len(g.Players) == 2
}

// Determines if the game board needs resetting
func (g *Game) shouldResetBoard() bool {
	return g.getWinnerSymbol() != noWinner || g.isFullBoard()
}

// Checks if it's the specified player's turn to play
func (g *Game) canPlayTurn(p Player) bool {
	return p.Connection == g.CurrentPlayer.Connection && g.isFullPlaces()
}

// Resets the game board to initial state
func (g *Game) resetBoard() {
	for i := 0; i < 9; i++ {
		g.Board[i] = fmt.Sprintf("%d", i)
	}
}

// Switches to the next player
func (g *Game) switchCurrentPlayer() {
	if g.CurrentPlayer.Index == 0 {
		g.CurrentPlayer = g.Players[1]
	} else {
		g.CurrentPlayer = g.Players[0]
	}
}

// Checks if a position on the board is free
func (g *Game) isFreePos(pos int) bool {
	return pos >= 0 && pos < 9 && g.Board[pos] != xSymbol && g.Board[pos] != oSymbol
}

// Marks a position on the board for the player's turn
func (g *Game) playTurn(p Player, pos int) {
	g.Board[pos] = p.Symbol
}

// Checks if there's a winner by verifying winning combinations
func (g *Game) getWinnerSymbol() string {
	for _, c := range winCombinations {
		if g.Board[c[0]] == g.Board[c[1]] && g.Board[c[1]] == g.Board[c[2]] {
			return g.Board[c[0]]
		}
	}
	return noWinner
}

// Returns the player corresponding to a specific symbol
func (g *Game) getPlayerWithSymbol(s string) *Player {
	if g.Players[0].Symbol == s {
		return &g.Players[0]
	}
	return &g.Players[1]
}

// Increments player's score
func (p *Player) incrementScore() {
	p.Score++
}

// Notifies players of the next turn
func (g *Game) dispatchNextTurn() {
	if !g.isFullPlaces() {
		return
	}
	for _, p := range g.Players {
		if p.Connection != g.CurrentPlayer.Connection {
			p.Connection.Write([]byte("\033[1;31mopponent's turn!\033[0m\n"))
		} else {
			p.Connection.Write([]byte("\033[1;32myour turn:\033[0m\n"))
		}
	}
}

// Broadcasts the current game state to all players
func (g *Game) dispatchGame() {
	for _, p := range g.Players {
		p.Connection.Write([]byte(g.String()))
	}
}

// Manages player input for the game connection
func handleGameConnection(g *Game, p Player) {
	scanner := bufio.NewScanner(p.Connection)
	defer handlePlayerQuit(g, p)

	for scanner.Scan() {
		pos, _ := strconv.Atoi(scanner.Text())
		handlePlayerPosition(pos, g, p)
	}
}

// Handles player disconnection and game reset
func handlePlayerQuit(g *Game, pl Player) {
	for i, p := range g.Players {
		if p.Connection == pl.Connection {
			g.Players = append(g.Players[:i], g.Players[i+1:]...)
			break
		}
	}
	pl.Connection.Close()
	g.resetBoard()
	g.Players[0].Score = 0
	g.Players[0].Connection.Write([]byte("\n\n Your opponent left the game.\n Reset.\n Waiting for a new opponent.\n"))
	fmt.Printf("client %v disconnected\n", pl.Connection.RemoteAddr().String())
}

// Handles player moves and game state updates
func handlePlayerPosition(pos int, g *Game, p Player) {
	if g.canPlayTurn(p) && g.isFreePos(pos) {
		g.playTurn(p, pos)
		g.switchCurrentPlayer()
	} else if !g.canPlayTurn(p) {
		p.Connection.Write([]byte("\033[1;33mIT'S NOT YOUR TURN\033[0m\n"))
		return
	} else if !g.isFreePos(pos) {
		p.Connection.Write([]byte("\033[1;31mTHIS PLACE IS ALREADY TAKEN\033[0m\n\n"))
		return
	}
	if ws := g.getWinnerSymbol(); ws != noWinner {
		winner := g.getPlayerWithSymbol(ws)
		winner.Connection.Write([]byte("\n\n\033[1;32mYOU WIN\033[0m\n\n\n"))
		loser := g.Players[1-winner.Index]
		loser.Connection.Write([]byte("\n\n\033[1;33mYOU LOSE\033[0m\n\n\n"))
		winner.incrementScore()
	}
	if g.shouldResetBoard() {
		g.resetBoard()
	}
	g.dispatchGame()
	g.dispatchNextTurn()
}

// Rejects a connection if the game is full
func rejectConnection(conn net.Conn) {
	fmt.Printf("rejecting client connected from %v\n", conn.RemoteAddr().String())
	conn.Write([]byte("\033[1;31mGame is full, try again later!\033[0m\n"))
	conn.Close()
}

// Main function to initialize server and accept connections
func main() {
	ip := "192.168.0.252"
	port := "8080"
	addr := ip + ":" + port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
		return
	}
	defer listener.Close()

	var game = Game{
		Board: make(map[int]string, 9),
	}
	game.resetBoard()

	fmt.Printf("Listening on %v \n\n", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error with accepting client: ", err)
			continue
		}

		if game.isFullPlaces() {
			rejectConnection(conn)
			continue
		}

		fmt.Printf("client connected from %v\n", conn.RemoteAddr().String())

		player := Player{
			Index:      len(game.Players),
			Connection: conn,
			Symbol:     []string{xSymbol, oSymbol}[len(game.Players)],
			Score:      0,
		}
		game.Players = append(game.Players, player)
		game.CurrentPlayer = player

		if game.isFullPlaces() {
			game.dispatchGame()
			game.dispatchNextTurn()
		}

		go handleGameConnection(&game, player)
	}
}
