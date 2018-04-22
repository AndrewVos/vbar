package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gotk3/gotk3/gtk"
)

// Block is the container class for the gtk.EventBox and gtk.Label.
type Block struct {
	AddBlock
	EventBox *gtk.EventBox
	Label    *gtk.Label
	Menu     *gtk.Menu
}

// Initialize builds widgets and sets up triggers.
func (b *Block) Initialize() error {
	err := b.initializeEventBox()
	if err != nil {
		return err
	}

	err = b.initializeLabel()
	if err != nil {
		return err
	}

	err = b.initializeCommand()
	if err != nil {
		return err
	}

	err = b.initializeClickCommand()
	if err != nil {
		return err
	}

	err = b.initializeTailCommand()
	if err != nil {
		return err
	}

	return nil
}

func (b *Block) initializeEventBox() error {
	return executeGtkSync(func() error {
		eventBox, err := gtk.EventBoxNew()
		b.EventBox = eventBox
		return err
	})
}

func (b *Block) initializeLabel() error {
	return executeGtkSync(func() error {
		label, err := gtk.LabelNew(b.Text)
		if err != nil {
			return err
		}
		b.Label = label
		b.EventBox.Add(label)
		err = applyClass(&label.Widget, "block")
		if err != nil {
			return err
		}
		err = applyClass(&label.Widget, b.Name)
		if err != nil {
			return err
		}
		return nil
	})
}

func (b *Block) initializeCommand() error {
	if b.Command == "" {
		return nil
	}

	b.startUpdatingLabel()

	if b.Interval != 0 {
		duration, _ := time.ParseDuration(fmt.Sprintf("%ds", b.Interval))
		tick := time.Tick(duration)
		go func() {
			for range tick {
				b.startUpdatingLabel()
			}
		}()
	}

	return nil
}

func (b *Block) initializeTailCommand() error {
	if b.TailCommand == "" {
		return nil
	}
	b.startUpdatingLabelForever()
	return nil
}

func (b *Block) initializeClickCommand() error {
	if b.ClickCommand == "" {
		return nil
	}

	return executeGtkSync(func() error {
		_, err := b.EventBox.Connect("button-release-event", func() {
			go func() {
				cmd := exec.Command("/bin/bash", "-c", b.ClickCommand)
				err := cmd.Run()
				if err != nil {
					log.Printf("ClickCommand finished with error: %v", err)
				}
			}()
		})
		return err
	})
}

func (b *Block) startUpdatingLabel() {
	go func() {
		cmd := exec.Command("/bin/bash", "-c", b.Command)
		cmd.Stderr = os.Stderr

		stdout, err := cmd.Output()
		if err == nil {
			b.setText(strings.TrimSpace(string(stdout)))
		} else {
			log.Printf("Command finished with error: %v", err)
			b.setText("ERROR")
		}
	}()
}

func (b *Block) setText(text string) {
	err := executeGtkSync(func() error {
		b.Label.SetText(text)
		return nil
	})
	if err != nil {
		log.Printf("Error setting text: %v", err)
	}
}

func (b *Block) startUpdatingLabelForever() {
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
