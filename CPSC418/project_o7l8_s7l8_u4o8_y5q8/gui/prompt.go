package gui

import (
	"../buffer"
	"github.com/nsf/termbox-go"
)

func redrawPrompt(prompt *buffer.Prompt, width, height int) {
	cursorX, cursorY, lines := prompt.GetDisplayInformation(width-1, height)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	drawLines(lines, height)
	if cursorY < height {
		termbox.SetCursor(cursorX, cursorY)
	}
	termbox.Flush()
}

func DrawPrompt(prompt *buffer.Prompt) error {
	width, height := termbox.Size()
	termbox.SetInputMode(termbox.InputEsc)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	redrawPrompt(prompt, width, height)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				appState.State = STATE_EXIT
				return nil
			case termbox.KeyCtrlK:
				termbox.Close()
				panic("Force kill")
			case termbox.KeyEsc:
				if appState.DocModel != nil {
					appState.State = STATE_DOCUMENT
					return nil
				}
			case termbox.KeyBackspace:
				prompt.Delete()
				redrawPrompt(prompt, width, height)
			case termbox.KeyBackspace2:
				prompt.Delete()
				redrawPrompt(prompt, width, height)
			case termbox.KeySpace:
				prompt.Insert(' ')
				redrawPrompt(prompt, width, height)
			case termbox.KeyTab:
				prompt.Insert('\t')
				redrawPrompt(prompt, width, height)
			case termbox.KeyEnter:
				doAction(prompt.ToString())
				return nil
			default:
				if ev.Key == 0 && ev.Ch <= 256 {
					prompt.Insert(byte(ev.Ch))
					redrawPrompt(prompt, width, height)
				}
			}
		case termbox.EventResize:
			width = ev.Width
			height = ev.Height
			redrawPrompt(prompt, width, height)
		case termbox.EventInterrupt:
			redrawPrompt(prompt, width, height)
		case termbox.EventError:
			termbox.Close()
			panic(ev.Err)
		}
	}
}
