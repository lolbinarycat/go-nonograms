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
	Colunms, Rows [][]int
}

func (data *NonogramData) CheckData() error {
	switch {
	case int32(len(data.Colunms)) != data.Width :
		return errors.New(fmt.Sprintf("json error: number of colunms (%d) is not equal to width (%d)",len(data.Colunms),data.Width))
	case int32(len(data.Rows)) != data.Height :
		return errors.New(fmt.Sprintf("json error: number of rows (%d) is not equal to height (%d)",len(data.Rows),data.Height))
	}

	return nil
}

var style = tcell.StyleDefault

func main() {
	nonogramFile, err := os.Open("test.json")
	ec(err)
	nonogramBytes, _ := ioutil.ReadAll(nonogramFile)

	var nonogramData NonogramData
	ec(nonogramData.CheckData())
	json.Unmarshal(nonogramBytes,&nonogramData)
	
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	screen, err := tcell.NewScreen()
	ec(err)
	err = screen.Init()
	ec(err)

	screen.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack))
	screen.Clear()

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
	
	drawNonogram(nonogramData,screen)	

	screen.Show()

loop:
	for {
		select {
		case <-quit:
			break loop
		}
	}

	screen.Fini()
	
}

func ec(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
	


func drawNonogram(non NonogramData,screen tcell.Screen) {

	i := int(non.Height)

	offset := 5
	
	for i > 0 {
		j := int(non.Width)
		for j > 0 {
			screen.SetCell(j+offset,i+offset,tcell.StyleDefault,'0')
			j--
		}
		i--
	}

	for i, _ := range non.Rows {
		for j , hint := range non.Rows[i] {
			hintStr := strconv.Itoa(hint)
			screen.SetCell(offset-len(non.Rows)+j,offset+i,style,rune(hintStr[0]))
		}
	}

	screen.Show()
}
		
