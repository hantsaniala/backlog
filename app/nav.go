package app

import "time"

type InputMode int

const (
	ModeNormal InputMode = iota
	ModeInsert
	ModeVisual
	ModeHelp
	ModeCommandPalette
)

func (m InputMode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeVisual:
		return "VISUAL"
	case ModeHelp:
		return "HELP"
	case ModeCommandPalette:
		return "CMD"
	default:
		return "NORMAL"
	}
}

func formatTime() string {
	return time.Now().Format("15:04")
}
