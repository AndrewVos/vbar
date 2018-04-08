package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gotk3/gotk3/gtk"
)

// Block is the container class for the gtk.EventBox and gtk.Label.
type Block struct {
	EventBox     *gtk.EventBox
	Label        *gtk.Label
	Menu         *gtk.Menu
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

func (b Block) updateLabel() {
	cmd := exec.Command("/bin/bash", "-c", b.Command)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.Output()
	if err == nil {
		b.Label.SetText(strings.TrimSpace(string(stdout)))
	} else {
		log.Printf("Command finished with error: %v", err)
		b.Label.SetText("ERROR")
	}
}

func (b Block) updateLabelForever() {
	go func() {
		cmd := exec.Command("/bin/bash", "-c", b.TailCommand)
		cmd.Stderr = os.Stderr

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("Couldn't get a stdout from command: %v", err)
			b.Label.SetText("ERROR")
			return
		}
		err = cmd.Start()
		if err != nil {
			log.Printf("Command finished with error: %v", err)
			b.Label.SetText("ERROR")
			return
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			b.Label.SetText(strings.TrimSpace(scanner.Text()))
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Couldn't read from command stdout: %v", err)
			b.Label.SetText("ERROR")
			return
		}
	}()
}
