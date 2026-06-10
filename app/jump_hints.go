package app

import "fmt"

type hintLabel string

func generateHints(count int) []hintLabel {
	if count <= 0 {
		return nil
	}
	letters := "abcdefghijklmnopqrstuvwxyz"
	var hints []hintLabel
	// 2-char hints: aa, ab, ..., ba, bb, ...
	for _, a := range letters {
		for _, b := range letters {
			hints = append(hints, hintLabel(fmt.Sprintf("%c%c", a, b)))
			if len(hints) >= count {
				return hints
			}
		}
	}
	return hints[:count]
}

type jumpHintState struct {
	active bool
	hints  map[int]hintLabel // row index in visibleRows -> hint label
	buffer string
}

func newJumpHintState() *jumpHintState {
	return &jumpHintState{
		active: false,
		hints:  make(map[int]hintLabel),
		buffer: "",
	}
}

func (j *jumpHintState) activate(count int) {
	j.active = true
	j.hints = make(map[int]hintLabel)
	j.buffer = ""
	labels := generateHints(count)
	for i, label := range labels {
		j.hints[i] = label
	}
}

func (j *jumpHintState) pushRune(r rune) (int, bool) {
	j.buffer += string(r)
	if len(j.buffer) >= 2 {
		target := j.matchHint(hintLabel(j.buffer))
		if target >= 0 {
			j.active = false
			j.buffer = ""
			return target, true
		}
		// No match, reset
		j.buffer = j.buffer[len(j.buffer)-1:]
		if len(j.buffer) >= 2 {
			j.active = false
			j.buffer = ""
			return -1, false
		}
	}
	return -1, false
}

func (j *jumpHintState) matchHint(label hintLabel) int {
	for rowIdx, hl := range j.hints {
		if hl == label {
			return rowIdx
		}
	}
	return -1
}

func (j *jumpHintState) cancel() {
	j.active = false
	j.hints = make(map[int]hintLabel)
	j.buffer = ""
}

// hintOverlay returns a label with the hint overlaid if this row has a hint.
func (j *jumpHintState) hintOverlay(rowIdx int, line string) string {
	if !j.active {
		return line
	}
	h, ok := j.hints[rowIdx]
	if !ok {
		return line
	}
	hintStr := jumpHintStyle.Render(string(h))
	// Render the hint as overlay. Use styled prefix + rest of line.
	// We prepend the hint label before the line content.
	return fmt.Sprintf("%s %s", hintStr, line)
}

func (j *jumpHintState) bufferDisplay() string {
	if !j.active {
		return ""
	}
	return fmt.Sprintf(" jump: %s_", j.buffer)
}
