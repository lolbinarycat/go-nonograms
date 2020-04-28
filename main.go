package main

import (
	"encoding/json"
	"fmt"
	"errors"
	"os"
	"io/ioutil"

	"github.com/gdamore/tcell"
	"strconv"
)

type NonogramData struct {
	Name string
	Width, Height int32
	Columns, Rows [][]int
}



func (data *NonogramData) CheckData() error {
	switch {
	case int32(len(data.Columns)) != data.Width :
		return errors.New(fmt.Sprintf("json error: number of columns (%d) is not equal to width (%d)",len(data.Columns),data.Width))
	case int32(len(data.Rows)) != data.Height :
		return errors.New(fmt.Sprintf("json error: number of rows (%d) is not equal to height (%d)",len(data.Rows),data.Height))
	}

	return nil
}

func MakeBoardState(data NonogramData) NonogramBoardState {
	//i := data.Width

	var nbs NonogramBoardState

	nbs = make([][]NonogramCellState,data.Width + 1)

	for i := range nbs {
		nbs[i] = make([]NonogramCellState,data.Height + 1)
		i--
	}
	return nbs
}

type NonogramBoardState [][]NonogramCellState
	

type NonogramCellState int

const (
	Empty NonogramCellState = iota + 1
	Filled
)


var style = tcell.StyleDefault

var stateRunes = [2]rune{ '0','\u2588' } //describes how to display board

func main() {

	
	
	nonogramFile, err := os.Open("test.json")
	ec(err)
	nonogramBytes, err := ioutil.ReadAll(nonogramFile)
	ec(err)

	var nonogramData NonogramData
	
	ec(json.Unmarshal(nonogramBytes,&nonogramData))
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
	
	var nonogramBoardState NonogramBoardState = MakeBoardState(nonogramData) //make([][]NonogramCellState,10)

	quit := make(chan struct{})
	go func() {
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
				default:
					switch ev.Rune(){
					case 'a':
						screen.SetCell(1,1,tcell.StyleDefault,'0')
						screen.Show()
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
	
	drawNonogram(nonogramData,nonogramBoardState,screen)	

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
	


func drawNonogram(non NonogramData,nonState NonogramBoardState,screen tcell.Screen) {

	defer screen.Show()

	
	i := int(non.Width)

	offset := 5
	
	for i > 0 {
		j := int(non.Height)
		for j > 0 {

			runeToDraw := 'e'
			
			if len(nonState) >= i && len(nonState[i]) >= j {
				runeToDraw = stateRunes[int(nonState[i][j])]
			} else {
				panic("error: nonogramBoardState is too short")
			}
			
			screen.SetCell(i+offset,j+offset,tcell.StyleDefault,runeToDraw)
			j--
		}
		i--
	}

	for i, _ := range non.Rows {
		for j , hint := range non.Rows[i] {
			hintStr := strconv.Itoa(hint)
			screen.SetCell(offset+j-len(non.Rows[i]),offset+i+1,style,rune(hintStr[0]))
		}
	}

	

	for i, _ := range non.Columns {
		for j , hint := range non.Columns[i] {
			hintStr := strconv.Itoa(hint)
			screen.SetCell(offset+i+1,offset+j-len(non.Columns[i]),style,rune(hintStr[0]))
		}
	}


	screen.Show()
}
		
/* func tryGetFromSlice(slice []interface{}, indexes ...int) (retrivedValue {
	defer func() {

	}()

	slice[indexes[0]]  */
