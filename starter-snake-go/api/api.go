package api

import (
	"encoding/json"
	"net/http"
)

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Snake struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Health int     `json:"health"`
	Body   []Coord `json:"body"`
}

type Board struct {
	Height int     `json:"height"`
	Width  int     `json:"width"`
	Food   []Coord `json:"food"`
	Snakes []Snake `json:"snakes"`
}

type Game struct {
	ID string `json:"id"`
}

type SnakeRequest struct {
	Game  Game  `json:"game"`
	Turn  int   `json:"turn"`
	Board Board `json:"board"`
	You   Snake `json:"you"`
}

type StartResponse struct {
	Color string `json:"color,omitempty"`
}

type MoveResponse struct {
	Move string `json:"move"`
}

func DecodeSnakeRequest(req *http.Request, decoded *SnakeRequest) error {
	err := json.NewDecoder(req.Body).Decode(&decoded)
	return err
}


const down = "down"
const up = "up"
const left = "left"
const right = "right"

var gameHeads = map[string]Coord{}
var PreviousEstimates = map[string]int{}

func GetMove(state SnakeRequest) string{

	//check for head or create if not existing

	if _, ok := gameHeads[state.Game.ID]; !ok{
		//print("new game making head ", state.You.Body[0], "\n")
		gameHeads[state.Game.ID] = state.You.Body[0]
	}

	head := gameHeads[state.Game.ID]

	occupiedPositions := make([][]int, state.Board.Width)
	for i := range occupiedPositions {
		occupiedPositions[i] = make([]int, state.Board.Height)
		for j := range occupiedPositions[i] {
			occupiedPositions[i][j] = -1
		}
	}

	headPositions := make([][]int, state.Board.Width)
	for i := range headPositions {
		headPositions[i] = make([]int, state.Board.Height)
		for j := range headPositions[i] {
			headPositions[i][j] = -1
		}
	}



	//each snake
	for i := 0; i < len(state.Board.Snakes); i++ {
		//each snake segment
		for j := 0; j < len(state.Board.Snakes[i].Body); j++{
			x := state.Board.Snakes[i].Body[j].X
			y := state.Board.Snakes[i].Body[j].Y
			if j == 0 && x != head.X && y != head.Y {
				headPositions[i][j] = i
			}else{
				occupiedPositions[x][y] = i
			}
		}
	}

	foodPositions := make([][]bool, state.Board.Width)
	for i := range foodPositions {
		foodPositions[i] = make([]bool, state.Board.Height)
		for j := range foodPositions[i] {
			foodPositions[i][j] = false
		}
	}

	for i := 0; i < len(state.Board.Food); i++ {
		foodPositions[state.Board.Food[i].X][state.Board.Food[i].Y] = true
	}


	//find non suicidal moves
	canMove := map[string]bool{
		down: true,
		up: true,
		left: true,
		right: true,
	}

	//notes 0 is top/left, boardsize is width/height-1
	canMove[down] = head.Y+1 <= (state.Board.Height-1) && occupiedPositions[head.X][head.Y+1] == -1
	canMove[up] = head.Y-1 >= 0 && occupiedPositions[head.X][head.Y-1] == -1

	canMove[left] = head.X-1 >= 0 && occupiedPositions[head.X-1][head.Y] == -1
	canMove[right] = head.X+1 <= (state.Board.Width-1) && occupiedPositions[head.X+1][head.Y] == -1

	estimatedValue := map[string]int{
		down: 0,
		up: 0,
		left: 0,
		right: 0,
	}

	channels := map[string]chan int{
		down: make(chan int, 1),
		up: make(chan int, 1),
		left: make(chan int, 1),
		right: make(chan int, 1),
	}

	if canMove[down]{
		go EstimateValue(channels[down], state, Coord{X: head.X, Y: head.Y+1}, head, occupiedPositions, foodPositions, 0)
	}else{
		go func(channel chan int) { channel <- IMMEDIATE_DEATH_VALUE }(channels[down])
	}

	if canMove[up]{
		go EstimateValue(channels[up], state, Coord{X: head.X, Y: head.Y-1}, head, occupiedPositions, foodPositions, 0)
	}else{
		go func(channel chan int) { channel <- IMMEDIATE_DEATH_VALUE }(channels[up])
	}

	if canMove[left]{
		go EstimateValue(channels[left], state, Coord{X: head.X-1, Y: head.Y}, head, occupiedPositions, foodPositions, 0)
	}else{
		go func(channel chan int) { channel <- IMMEDIATE_DEATH_VALUE }(channels[left])
	}

	if canMove[right]{
		go EstimateValue(channels[right], state, Coord{X: head.X+1, Y: head.Y}, head, occupiedPositions, foodPositions, 0)
	}else{
		go func(channel chan int) { channel <- IMMEDIATE_DEATH_VALUE }(channels[right])
	}

	estimatedValue[down] = <- channels[down]
	estimatedValue[up] = <- channels[up]
	estimatedValue[left] = <- channels[left]
	estimatedValue[right] = <- channels[right]


	maxIndex := MapMaxIndex(estimatedValue)

	//PrintBoard(occupiedPositions, foodPositions)
	println("estimates - down: ",estimatedValue[down],"  up:  ",estimatedValue[up],"  left:  ",estimatedValue[left],"  right:  ",estimatedValue[right])
	println("moving = ", maxIndex)

	PreviousEstimates = estimatedValue

	if maxIndex == down{
		head.Y = head.Y+1
		gameHeads[state.Game.ID] = head
		return down
	}else if maxIndex == up{
		head.Y = head.Y-1
		gameHeads[state.Game.ID] = head
		return up
	}else if maxIndex == left{
		head.X = head.X-1
		gameHeads[state.Game.ID] = head
		return left
	}else{
		head.X = head.X+1
		gameHeads[state.Game.ID] = head
		return right
	}
}

func MapMaxIndex(vals map[string]int) string {

	var maxIndex string
	max := -1

	for key, val := range vals{

		if max == -1{
			max = val
			maxIndex = key
		}else if val > max{
			max = val
			maxIndex = key
		}

	}

	return maxIndex

}

func PrintBoard(occupiedPositions [][]int, foodPositions [][]bool){
	println("Board State:")

	rows := make([]string, len(occupiedPositions[0]))


	for i := 0; i < len(occupiedPositions); i++{

		for j := 0; j < len(occupiedPositions[i]); j++{
			rows[j] = rows[j] + "|"

			if occupiedPositions[i][j] != -1{
				rows[j] = rows[j] + "S"
			}
			if foodPositions[i][j]{
				rows[j] = rows[j] + "F"
			}
		}
	}

	for i := 0; i < len(rows); i++{
		println(rows[i])
	}
}

func EstimateValue(channel chan int, state SnakeRequest, spot Coord, head Coord, occupiedPositions [][]int, foodPositions [][]bool, recursion int) {

	relevanceFactor := 1 - (recursion/RECURSION_LIMIT)

	//only called if move is valid
	value := BASE_VALUE * relevanceFactor


	//if food space
	if foodPositions[spot.X][spot.Y]{
		if state.You.Health > HEALTH_THRESHOLD && foodPositions[spot.X][spot.Y]{
			value = value + ( FOOD_VALUE  * relevanceFactor )
		}else {
			value = value + (LOW_FOOD_VALUE * relevanceFactor)
		}
	}

	//if occupied space
	if occupiedPositions[spot.X][spot.Y] != -1{
		//but can kill them (note that this can't be you as you are ==)
		if len(state.You.Body)<len(state.Board.Snakes[occupiedPositions[spot.X][spot.Y]].Body){

			//is another snake that is smaller than you check if it's a kill
			if spot.X == state.Board.Snakes[occupiedPositions[spot.X][spot.Y]].Body[0].X && spot.Y == state.Board.Snakes[occupiedPositions[spot.X][spot.Y]].Body[0].Y {
				//we don't want to track head killing outside of this because they will move
				if (recursion == 0) {
					value = value + (KILL_VALUE * relevanceFactor)
				}
			}else{
				value = value + ( DEATH_VALUE  * relevanceFactor )
			}

		}else{
			if (recursion == 0){
				value = value + IMMEDIATE_DEATH_VALUE
			}else {
				value = value + (DEATH_VALUE * relevanceFactor)
			}
		}
	}

	//if space is next to a wall
	if spot.X == 0 || spot.X == (state.Board.Width-1){
		value = value +  ( NEXT_TO_WALL_VALUE * relevanceFactor )
	}

	if spot.Y == 0 || spot.Y == (state.Board.Height-1){
		value = value + ( NEXT_TO_WALL_VALUE  * relevanceFactor )
	}

	if recursion < RECURSION_LIMIT {
		//check next space
		//find non suicidal moves
		canMove := map[string]bool{
			down: true,
			up: true,
			left: true,
			right: true,
		}

		//notes 0 is top/left, boardsize is width/height-1
		canMove[down] = spot.Y+1 <= (state.Board.Height-1) && occupiedPositions[spot.X][spot.Y+1] == -1
		canMove[up] = spot.Y-1 >= 0 && occupiedPositions[spot.X][spot.Y-1] == -1

		canMove[left] = spot.X-1 >= 0 && occupiedPositions[spot.X-1][spot.Y] == -1
		canMove[right] = spot.X+1 <= (state.Board.Width-1) && occupiedPositions[spot.X+1][spot.Y] == -1

		removeTail := !foodPositions[spot.X][spot.Y]
		tailPos := len(state.You.Body)-1
		if (removeTail) {
			occupiedPositions[state.You.Body[tailPos].X][state.You.Body[tailPos].Y] = -1
		}


		recursion = recursion+1
		channels := map[string]chan int{
			down: make(chan int, 1),
			up: make(chan int, 1),
			left: make(chan int, 1),
			right: make(chan int, 1),
		}


		failValue := 0
		for i := recursion; i < RECURSION_LIMIT; i++{
			rf :=  1 - (recursion/RECURSION_LIMIT)
			failValue = failValue + (DEATH_VALUE * rf)
		}


		if canMove[down]{
			duplicate := make([][]int, len(occupiedPositions))
			for i := range occupiedPositions {
				duplicate[i] = make([]int, len(occupiedPositions[i]))
				copy(duplicate[i], occupiedPositions[i])
			}
			duplicate[spot.X][spot.Y]=0

			foodDupe := make([][]bool, len(foodPositions))
			for i := range foodPositions {
				foodDupe[i] = make([]bool, len(foodPositions[i]))
				copy(foodDupe[i], foodPositions[i])
			}
			go EstimateValue(channels[down], state, Coord{X: spot.X, Y: spot.Y+1}, spot, duplicate, foodDupe, recursion)
		}else{
			go func(channel chan int) { channel <- failValue }(channels[down])
		}

		if canMove[up]{
			duplicate := make([][]int, len(occupiedPositions))
			for i := range occupiedPositions {
				duplicate[i] = make([]int, len(occupiedPositions[i]))
				copy(duplicate[i], occupiedPositions[i])
			}
			duplicate[spot.X][spot.Y]=0

			foodDupe := make([][]bool, len(foodPositions))
			for i := range foodPositions {
				foodDupe[i] = make([]bool, len(foodPositions[i]))
				copy(foodDupe[i], foodPositions[i])
			}
			go EstimateValue(channels[up], state, Coord{X: spot.X, Y: spot.Y-1}, spot, duplicate, foodDupe, recursion)
		}else{
			go func(channel chan int) { channel <- failValue }(channels[up])
		}

		if canMove[left]{
			duplicate := make([][]int, len(occupiedPositions))
			for i := range occupiedPositions {
				duplicate[i] = make([]int, len(occupiedPositions[i]))
				copy(duplicate[i], occupiedPositions[i])
			}
			duplicate[spot.X][spot.Y]=0

			foodDupe := make([][]bool, len(foodPositions))
			for i := range foodPositions {
				foodDupe[i] = make([]bool, len(foodPositions[i]))
				copy(foodDupe[i], foodPositions[i])
			}
			go EstimateValue(channels[left], state, Coord{X: spot.X-1, Y: spot.Y}, spot, duplicate, foodDupe, recursion)
		}else{
			go func(channel chan int) { channel <- failValue }(channels[left])
		}

		if canMove[right]{
			duplicate := make([][]int, len(occupiedPositions))
			for i := range occupiedPositions {
				duplicate[i] = make([]int, len(occupiedPositions[i]))
				copy(duplicate[i], occupiedPositions[i])
			}
			duplicate[spot.X][spot.Y]=0

			foodDupe := make([][]bool, len(foodPositions))
			for i := range foodPositions {
				foodDupe[i] = make([]bool, len(foodPositions[i]))
				copy(foodDupe[i], foodPositions[i])
			}
			go EstimateValue(channels[right], state, Coord{X: spot.X+1, Y: spot.Y}, spot, duplicate, foodDupe, recursion)
		}else{
			go func(channel chan int) { channel <- failValue }(channels[right])
		}

		downV := <- channels[down]
		upV := <- channels[up]
		leftV := <- channels[left]
		rightV := <- channels[right]

		value += downV
		value += upV
		value += leftV
		value += rightV
	}


	channel <- value
}