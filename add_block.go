package main

// AddBlock contains the arguments used for the add-block command.
type AddBlock struct {
	Name         string
	Text         string
	Left         bool
	Center       bool
	Right        bool
	Command      string
	TailCommand  string
	Interval     int
	ClickCommand string
}
