package buffer

type Prompt struct {
	text  []byte
	input []byte
}

func NewPrompt(str string) *Prompt {
	return &Prompt{text: []byte(str), input: make([]byte, 0, 4)}
}

func (prompt *Prompt) Insert(ch byte) {
	prompt.input = append(prompt.input, ch)
}

func (prompt *Prompt) Delete() {
	if len(prompt.input) > 0 {
		prompt.input = prompt.input[:len(prompt.input)-1]
	}
}

// no wrapping
func (prompt *Prompt) GetDisplayInformation(width, height int) (int, int, *Line) {
	sentinel := &Line{}
	line := sentinel
	textHeight := 0
	i := 0
	last := 0
	for i < len(prompt.text) {
		lineWidth := 0
		for i < len(prompt.text) {
			if prompt.text[i] == '\n' {
				i++
				break
			}
			if lineWidth+charLength(prompt.text[i]) > width {
				break
			}
			lineWidth += charLength(prompt.text[i])
			i++
		}
		line.Next = &Line{Bytes: prompt.text[last:i], Prev: line}
		line = line.Next
		textHeight++
		last = i
	}

	i = 0
	last = 0
	inputHeight := 0
	for i < len(prompt.input) {
		lineWidth := 0
		for i < len(prompt.input) {
			if lineWidth+charLength(prompt.input[i]) > width {
				break
			}
			lineWidth += charLength(prompt.input[i])
			i++
		}
		line.Next = &Line{Bytes: prompt.input[last:i], Prev: line}
		line = line.Next
		inputHeight++
		last = i
	}

	if inputHeight == 0 {
		line.Next = &Line{Bytes: make([]byte, 0, width+1), Prev: line}
		line = line.Next
		inputHeight++
	}

	if sentinel.Next == nil {
		return 0, 0, nil
	} else {
		sentinel.Next.Prev = nil
		return sliceLength(line.Bytes), textHeight + inputHeight - 1, sentinel.Next
	}
}

func (prompt *Prompt) ToString() string {
	return string(prompt.input)
}
