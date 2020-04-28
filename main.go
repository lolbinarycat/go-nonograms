package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gdamore/tcell"
	"strconv"
)

type NonogramData struct {
	Name          string
	Width, Height int32
	Columns, Rows [][]int
}

func (data *NonogramData) CheckData() error {
	switch {
	case int32(len(data.Columns)) != data.Width:
		return errors.New(fmt.Sprintf("json error: number of columns (%d) is not equal to width (%d)", len(data.Columns), data.Width))
	case int32(len(data.Rows)) != data.Height:
		return errors.New(fmt.Sprintf("json error: number of rows (%d) is not equal to height (%d)", len(data.Rows), data.Height))
	}

	return nil
}

func MakeBoardState(data NonogramData) NonogramBoardState {
	//i := data.Width

	var nbs NonogramBoardState

	nbs = make([][]NonogramCellState, data.Width+1)

	for i := range nbs {
		nbs[i] = make([]NonogramCellState, data.Height+1)
		i--
	}
	return nbs
}

type NonogramBoardState [][]NonogramCellState

func (cell *NonogramCellState) Cycle() {
	defer func() {
		if recover() != nil {
			*cell = 0
		}
	}()
	if *cell + 2 == EndState {
		*cell = 0
	} else {
		*cell++
	}
}
	

type NonogramCellState int

const (
	Empty NonogramCellState = iota + 1
	Filled
	EndState //this is here for logic reasons, it is not an actual state
)

type Direction int

const (
	Up Direction = iota + 1
	Down
	Left
	Right
)

var style = tcell.StyleDefault

var stateRunes = [...]rune{'0', '\u2588','e'} //describes how to display board

var (
	cursorRune = '\u25A3'
	cursorPos  = [2]int{0, 0}
)

func main() {

	nonogramFile, err := os.Open("test.json")
	ec(err)
	nonogramBytes, err := ioutil.ReadAll(nonogramFile)
	ec(err)

	var nonogramData NonogramData

	ec(json.Unmarshal(nonogramBytes, &nonogramData))
	ec(nonogramData.CheckData())

	//fmt.Println(nonogramData)
	//os.Exit(53)

	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	screen, err := tcell.NewScreen()
	ec(err)
	err = screen.Init()
	ec(err)

	defer screen.Fini()

	screen.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack))
	screen.Clear()

	//fmt.Println("\a")

	var nonState NonogramBoardState = MakeBoardState(nonogramData) //make([][]NonogramCellState,10)

	upScrn := func() { //update the screen
		drawNonogram(nonogramData, nonState, screen)
		screen.Show()
	}

	quit := make(chan struct{})
	go func() {
		defer screen.Fini()
		for {
			ev := screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(quit)
					return
				case tcell.KeyCtrlL:
					screen.Sync()
				case tcell.KeyUp:
					moveCursor(Up, 1, nonState)
					upScrn()
				case tcell.KeyDown:
					moveCursor(Down, 1, nonState)
					drawNonogram(nonogramData, nonState, screen)
					screen.Show()
				case tcell.KeyLeft:
					moveCursor(Left, 1, nonState)
					drawNonogram(nonogramData, nonState, screen)
					screen.Show()
				case tcell.KeyRight:
					moveCursor(Right, 1, nonState)
					drawNonogram(nonogramData, nonState, screen)
					screen.Show()
					
				default:
					switch ev.Rune() {
					case 'a':
						screen.SetCell(1, 1, tcell.StyleDefault, '0')
						screen.Show()
					case ' ':
						nonState[cursorPos[0]+1][cursorPos[1]+1].Cycle()
						//upScrn()
					case 'q':
						close(quit)
						return
					}
				}

			case *tcell.EventResize:
				screen.Sync()
			}
		}
	}()

	drawNonogram(nonogramData, nonState, screen)

	screen.Show()

loop:
	for {
		select {
		case <-quit:
			break loop
		}
	}

}

func ec(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func drawNonogram(non NonogramData, nonState NonogramBoardState, screen tcell.Screen) {

	defer screen.Show()

	i := int(non.Width)

	offset := 5

	for i > 0 {
		j := int(non.Height)
		for j > 0 {

			runeToDraw := 'e' //if this shows up, something has gone wrong.

			if len(nonState) >= i && len(nonState[i]) >= j {
				runeToDraw = stateRunes[int(nonState[i][j])]
			} else {
				panic("error: nonState is too short")
			}

			screen.SetCell(i+offset, j+offset, tcell.StyleDefault, runeToDraw)
			j--
		}
		i--
	}

	for i, _ := range non.Rows {
		for j, hint := range non.Rows[i] {
			hintStr := strconv.Itoa(hint)
			screen.SetCell(offset+j-len(non.Rows[i]), offset+i+1, style, rune(hintStr[0]))
		}
	}

	for i, _ := range non.Columns {
		for j, hint := range non.Columns[i] {
			hintStr := strconv.Itoa(hint)
			screen.SetCell(offset+i+1, offset+j-len(non.Columns[i]), style, rune(hintStr[0]))
		}
	}

	screen.SetCell(cursorPos[0]+offset+1,cursorPos[1]+offset+1, style, cursorRune)
		

	screen.Show()
}

func moveCursor(dir Direction, dist int, nonState NonogramBoardState) (moveCompleted bool) {
	cursorPosBuffer := cursorPos
	cursorMax := [2]int{len(nonState), len(nonState[0])}

	switch dir {
	case Up:
		cursorPosBuffer[1] -= dist //because origin is at top left, going up involves decreacing the y value
	case Down:
		cursorPosBuffer[1] += dist
	case Left:
		cursorPosBuffer[0] -= dist
	case Right:
		cursorPosBuffer[0] += dist
	default:
		panic(fmt.Sprintf("moveCursor: %d is not recognized as a direction", dir))
	}
	if cursorPosBuffer[0] <= cursorMax[0] && cursorPosBuffer[0] >= 0 && cursorPosBuffer[1] <= cursorMax[1] && cursorPosBuffer[1] >= 0 { //this massive statement checks if the cursor would be in a valid position
		cursorPos = cursorPosBuffer
		return true
	} else {
		return false
	}

}

/* func tryGetFromSlice(slice []interface{}, indexes ...int) (retrivedValue {
defer func() {

}()

slice[indexes[0]]  */
