package main

import "C"

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

// Window is the container for the panel
type Window struct {
	gtkWindow       *gtk.Window
	gtkPanel        *gtk.Grid
	lastLeftBlock   *gtk.EventBox
	lastCenterBlock *gtk.EventBox
	lastRightBlock  *gtk.EventBox
	blocks          []*blockOptions
}

// WindowNew creates a new Window
func WindowNew() (*Window, error) {
	var window = &Window{}

	gtkWindow, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return nil, err
	}
	window.gtkWindow = gtkWindow

	window.gtkWindow.SetAppPaintable(true)
	window.gtkWindow.SetDecorated(false)
	window.gtkWindow.SetResizable(false)
	window.gtkWindow.SetSkipPagerHint(true)
	window.gtkWindow.SetSkipTaskbarHint(true)
	window.gtkWindow.SetTypeHint(gdk.WINDOW_TYPE_HINT_DOCK)
	window.gtkWindow.SetVExpand(false)
	window.gtkWindow.SetPosition(gtk.WIN_POS_NONE)
	window.gtkWindow.Move(0, 0)
	window.gtkWindow.SetSizeRequest(-1, -1)

	window.gtkWindow.Connect("destroy", func() {
		gtk.MainQuit()
	})

	window.gtkWindow.Connect("realize", func() {
		window.gtkWindow.ShowAll()
		updateDimensions(window.gtkWindow, &window.gtkPanel.Widget)
	})

	gtkPanel, err := gtk.GridNew()
	if err != nil {
		log.Fatal(err)
	}
	window.gtkPanel = gtkPanel

	applyClass(&window.gtkPanel.Widget, "panel")
	window.gtkWindow.Add(window.gtkPanel)

	enableTransparency(window.gtkWindow)

	return window, nil
}

func (w *Window) addBlock(options *blockOptions) {
	w.blocks = append(w.blocks, options)

	eventBox, err := gtk.EventBoxNew()
	if err != nil {
		log.Println(err)
		return
	}
	options.EventBox = eventBox

	label, err := gtk.LabelNew(options.Text)
	if err != nil {
		log.Println(err)
		return
	}
	applyClass(&label.Widget, "block")
	applyClass(&label.Widget, options.Name)
	options.Label = label
	eventBox.Add(label)

	if options.Left {
		w.addBlockLeft(eventBox)
	} else if options.Center {
		w.addBlockCenter(eventBox)
	} else if options.Right {
		w.addBlockRight(eventBox)
	}

	if options.Command != "" {
		if options.Name == "title" {
			os.Exit(1)
		}
		options.updateLabel()

		if options.Interval != 0 {
			duration, _ := time.ParseDuration(fmt.Sprintf("%ds", options.Interval))
			tick := time.Tick(duration)
			go func() {
				for range tick {
					options.updateLabel()
				}
			}()
		}
	} else if options.TailCommand != "" {
		options.updateLabelForever()
	}

	if options.ClickCommand != "" {
		options.EventBox.Connect("button-release-event", func() {
			cmd := exec.Command("/bin/bash", "-c", options.ClickCommand)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err = cmd.Run()
			if err != nil {
				log.Printf("Command finished with error: %v", err)
			}
		})
	}

	window.gtkWindow.ShowAll()
}

func (w *Window) addMenu(options menuOptions) error {
	for _, block := range w.blocks {
		if block.Name == options.Name {
			if block.Menu == nil {
				menu, err := gtk.MenuNew()
				if err != nil {
					log.Fatal(err)
				}
				block.Menu = menu

				applyClass(&block.Menu.Widget, "menu")

				block.EventBox.Connect("button-release-event", func() {
					popupMenuAt(&block.EventBox.Widget, block.Menu)
				})
			}

			menuItem, err := gtk.MenuItemNewWithLabel(options.Text)
			if err != nil {
				log.Fatal(err)
			}
			menuItem.Connect("activate", func() {
				cmd := exec.Command("/bin/bash", "-c", options.Command)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err = cmd.Run()
				if err != nil {
					log.Printf("Command finished with error: %v", err)
				}
			})
			block.Menu.Add(menuItem)
			block.Menu.ShowAll()
		}
	}

	return nil
}

func (w *Window) updateBlock(options updateOptions) {
	for _, block := range w.blocks {
		if block.Name == options.Name {
			block.updateLabel()
			break
		}
	}
}

func (w *Window) addBlockLeft(block *gtk.EventBox) {
	block.SetHAlign(gtk.ALIGN_START)

	if w.lastLeftBlock != nil {
		w.gtkPanel.AttachNextTo(block, w.lastLeftBlock, gtk.POS_RIGHT, 1, 1)
	} else if w.lastCenterBlock != nil {
		w.gtkPanel.AttachNextTo(block, w.lastCenterBlock, gtk.POS_LEFT, 1, 1)
	} else if w.lastRightBlock != nil {
		w.gtkPanel.AttachNextTo(block, w.lastRightBlock, gtk.POS_LEFT, 1, 1)
	} else {
		w.gtkPanel.Attach(block, 0, 0, 1, 1)
	}
	w.lastLeftBlock = block
}

func (w *Window) addBlockCenter(block *gtk.EventBox) {
	block.SetHAlign(gtk.ALIGN_CENTER)
	block.SetHExpand(true)

	if w.lastCenterBlock != nil {
		w.gtkPanel.AttachNextTo(block, w.lastCenterBlock, gtk.POS_RIGHT, 1, 1)
	} else if w.lastLeftBlock != nil {
		w.gtkPanel.AttachNextTo(block, w.lastLeftBlock, gtk.POS_RIGHT, 1, 1)
	} else if w.lastRightBlock != nil {
		w.gtkPanel.AttachNextTo(block, w.lastRightBlock, gtk.POS_LEFT, 1, 1)
	} else {
		w.gtkPanel.Attach(block, 0, 0, 1, 1)
	}
	w.lastCenterBlock = block

}

func (w *Window) addBlockRight(block *gtk.EventBox) {
	block.SetHAlign(gtk.ALIGN_END)

	if w.lastRightBlock != nil {
		w.gtkPanel.AttachNextTo(block, w.lastRightBlock, gtk.POS_RIGHT, 1, 1)
	} else if w.lastCenterBlock != nil {
		w.gtkPanel.AttachNextTo(block, w.lastCenterBlock, gtk.POS_RIGHT, 1, 1)
	} else if w.lastLeftBlock != nil {
		w.gtkPanel.AttachNextTo(block, w.lastLeftBlock, gtk.POS_RIGHT, 1, 1)
	} else {
		w.gtkPanel.Attach(block, 0, 0, 1, 1)
	}
	w.lastRightBlock = block
}
