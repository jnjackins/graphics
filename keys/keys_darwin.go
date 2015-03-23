package keys

const (
	Backspace = 8
	Return    = 10
	Escape    = 27
	Down      = 128

	// TODO: is this a bug coming from plan9port, or from the plan9 Go package?
	Up    = 61454
	Left  = 61457
	Right = 61458

	Copy  = 61795 // cmd-c
	Save  = 61811 // cmd-s
	Paste = 61814 // cmd-v
	Cut   = 61816 // cmd-x
	Redo  = 61817 // cmd-y
	Undo  = 61818 // cmd-z
)
