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
	blocks          []*Block
	cssApplier      *CSSApplier
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

func (w *Window) addBlock(block *Block) error {
	w.blocks = append(w.blocks, block)

	eventBox, err := gtk.EventBoxNew()
	if err != nil {
		return err
	}
	block.EventBox = eventBox

	label, err := gtk.LabelNew(block.Text)
	if err != nil {
		return err
	}
	applyClass(&label.Widget, "block")
	applyClass(&label.Widget, block.Name)
	block.Label = label
	eventBox.Add(label)

	if block.Left {
		w.addBlockLeft(eventBox)
	} else if block.Center {
		w.addBlockCenter(eventBox)
	} else if block.Right {
		w.addBlockRight(eventBox)
	}

	if block.Command != "" {
		if block.Name == "title" {
			os.Exit(1)
		}
		block.updateLabel()

		if block.Interval != 0 {
			duration, _ := time.ParseDuration(fmt.Sprintf("%ds", block.Interval))
			tick := time.Tick(duration)
			go func() {
				for range tick {
					block.updateLabel()
				}
			}()
		}
	} else if block.TailCommand != "" {
		block.updateLabelForever()
	}

	if block.ClickCommand != "" {
		block.EventBox.Connect("button-release-event", func() {
			cmd := exec.Command("/bin/bash", "-c", block.ClickCommand)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err = cmd.Run()
			if err != nil {
				log.Printf("Command finished with error: %v", err)
			}
		})
	}

	window.gtkWindow.ShowAll()

	return nil
}

func (w *Window) applyCSS(addCSS AddCSS) error {
	if w.cssApplier == nil {
		w.cssApplier = &CSSApplier{}
	}

	screen, err := window.gtkWindow.GetScreen()
	if err != nil {
		return err
	}

	err = w.cssApplier.Apply(screen, addCSS)
	if err != nil {
		return err
	}

	return nil
}

func (w *Window) addMenu(addMenu AddMenu) error {
	for _, block := range w.blocks {
		if block.Name == addMenu.Name {
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

			menuItem, err := gtk.MenuItemNewWithLabel(addMenu.Text)
			if err != nil {
				log.Fatal(err)
			}
			menuItem.Connect("activate", func() {
				cmd := exec.Command("/bin/bash", "-c", addMenu.Command)
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

func (w *Window) updateBlock(update Update) {
	for _, block := range w.blocks {
		if block.Name == update.Name {
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
