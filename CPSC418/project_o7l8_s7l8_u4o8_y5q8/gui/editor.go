package gui

import (
	"github.com/nsf/termbox-go"
)

func redrawEditor(screenY, height int) int {
	screenY, cursorX, cursorY, lines := appState.DocModel.Buffer.GetDisplayInformation(screenY, height)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	drawLines(lines, height)
	termbox.SetCursor(cursorX, cursorY)
	termbox.Flush()
	return screenY
}

func DrawEditor() {
	width, height := termbox.Size()
	termbox.SetInputMode(termbox.InputEsc)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	docModel := appState.DocModel
	docModel.Buffer.Resize(width - 1)
	appState.ScreenY = redrawEditor(appState.ScreenY, height)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				appState.State = STATE_EXIT
				return
			case termbox.KeyCtrlK:
				termbox.Close()
				panic("Force kill")
			case termbox.KeyEsc:
				appState.State = STATE_MENU
				return
			case termbox.KeyArrowLeft:
				docModel.Buffer.MoveLeft()
				appState.ScreenY = redrawEditor(appState.ScreenY, height)
			case termbox.KeyArrowRight:
				docModel.Buffer.MoveRight()
				appState.ScreenY = redrawEditor(appState.ScreenY, height)
			case termbox.KeyArrowUp:
				docModel.Buffer.MoveUp()
				appState.ScreenY = redrawEditor(appState.ScreenY, height)
			case termbox.KeyArrowDown:
				docModel.Buffer.MoveDown()
				appState.ScreenY = redrawEditor(appState.ScreenY, height)
			case termbox.KeyBackspace:
				docModel.LocalBackspace()
				appState.ScreenY = redrawEditor(appState.ScreenY, height)
			case termbox.KeyBackspace2:
				docModel.LocalBackspace()
				appState.ScreenY = redrawEditor(appState.ScreenY, height)
			case termbox.KeyDelete:
				docModel.LocalDelete()
				appState.ScreenY = redrawEditor(appState.ScreenY, height)
			case termbox.KeySpace:
				docModel.LocalInsert(' ')
				appState.ScreenY = redrawEditor(appState.ScreenY, height)
			case termbox.KeyTab:
				docModel.LocalInsert('\t')
				appState.ScreenY = redrawEditor(appState.ScreenY, height)
			case termbox.KeyEnter:
				docModel.LocalInsert('\n')
				appState.ScreenY = redrawEditor(appState.ScreenY, height)
			default:
				if ev.Key == 0 && ev.Ch <= 256 {
					docModel.LocalInsert(byte(ev.Ch))
					appState.ScreenY = redrawEditor(appState.ScreenY, height)
				}
			}
		case termbox.EventResize:
			width = ev.Width
			height = ev.Height
			docModel.Buffer.Resize(width - 1)
			// this is a bug within the library, without this call clear would panic when
			// cursor is outside of resized window
			termbox.HideCursor()
			appState.ScreenY = redrawEditor(appState.ScreenY, height)
		case termbox.EventInterrupt:
			// remote operations
			appState.ScreenY = redrawEditor(appState.ScreenY, height)
		case termbox.EventError:
			termbox.Close()
			panic(ev.Err)
		}
	}
}
