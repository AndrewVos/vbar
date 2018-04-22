package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// Block is the container class for the gtk.EventBox and gtk.Label.
type Block struct {
	AddBlock
	EventBox *gtk.EventBox
	Label    *gtk.Label
	Menu     *gtk.Menu
}

func (b *Block) updateLabel() {
	cmd := exec.Command("/bin/bash", "-c", b.Command)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.Output()
	if err == nil {
		b.setText(strings.TrimSpace(string(stdout)))
	} else {
		log.Printf("Command finished with error: %v", err)
		b.setText("ERROR")
	}
}

func (b *Block) setText(text string) {
	_, err := glib.IdleAdd(func(text string) {
		b.Label.SetText(text)
	}, text)
	if err != nil {
		log.Panicf("Error setting text: %v", err)
	}
}

func (b *Block) updateLabelForever() {
	go func() {
		cmd := exec.Command("/bin/bash", "-c", b.TailCommand)
		cmd.Stderr = os.Stderr

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("Couldn't get a stdout from command: %v", err)
			b.setText("ERROR")
			return
		}
		err = cmd.Start()
		if err != nil {
			log.Printf("TailCommand finished with error: %v", err)
			b.setText("ERROR")
			return
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			b.setText(strings.TrimSpace(scanner.Text()))
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Couldn't read from command stdout: %v", err)
			b.setText("ERROR")
			return
		}
	}()
}
