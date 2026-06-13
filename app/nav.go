package app

type InputMode int

const (
	ModeNormal InputMode = iota
	ModeHelp
	ModeCommandPalette
)

func (m InputMode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeHelp:
		return "HELP"
	case ModeCommandPalette:
		return "CMD"
	default:
		return "NORMAL"
	}
}
