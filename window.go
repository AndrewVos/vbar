package main

import "C"

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
)

// Window is the container for the bar
type Window struct {
	gtkWindow       *gtk.Window
	gtkBar          *gtk.Grid
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
		updateDimensions(window.gtkWindow, &window.gtkBar.Widget)
	})

	gtkBar, err := gtk.GridNew()
	if err != nil {
		log.Fatal(err)
	}
	window.gtkBar = gtkBar

	window.gtkWindow.Add(window.gtkBar)

	err = applyClass(&window.gtkBar.Widget, "bar")
	if err != nil {
		return nil, err
	}

	enableTransparency(window.gtkWindow)

	return window, nil
}

func (w *Window) addBlock(addBlock AddBlock) error {
	block := &Block{AddBlock: addBlock}
	w.blocks = append(w.blocks, block)

	err := block.Initialize()
	if err != nil {
		return err
	}

	err = executeGtkSync(func() error {
		if block.Left {
			w.addBlockLeft(block)
		} else if block.Center {
			w.addBlockCenter(block)
		} else if block.Right {
			w.addBlockRight(block)
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = executeGtkSync(func() error {
		window.gtkWindow.ShowAll()
		return nil
	})

	return err
}

func (w *Window) addCSS(addCSS AddCSS) error {
	if w.cssApplier == nil {
		w.cssApplier = &CSSApplier{}
	}

	return executeGtkSync(func() error {
		screen, err := window.gtkWindow.GetScreen()
		if err != nil {
			return err
		}

		err = w.cssApplier.Apply(screen, addCSS)
		if err != nil {
			return err
		}

		return nil
	})
}

func (w *Window) addMenu(addMenu AddMenu) error {
	block := w.findBlock(addMenu.Name)
	if block == nil {
		return fmt.Errorf("couldn't find block %s", addMenu.Name)
	}

	if block.Menu == nil {
		err := executeGtkSync(func() error {
			menu, err := gtk.MenuNew()
			if err != nil {
				return err
			}
			block.Menu = menu

			err = applyClass(&block.Menu.Widget, "menu")
			if err != nil {
				return err
			}

			_, err = block.EventBox.Connect("button-release-event", func() {
				popupMenuAt(&block.EventBox.Widget, block.Menu)
			})
			return err
		})
		if err != nil {
			return err
		}
	}

	return executeGtkSync(func() error {
		menuItem, err := gtk.MenuItemNewWithLabel(addMenu.Text)
		if err != nil {
			return err
		}
		menuItem.Connect("activate", func() {
			cmd := exec.Command("/bin/bash", "-c", addMenu.Command)
			err = cmd.Run()
			if err != nil {
				log.Printf("Command finished with error: %v", err)
			}
		})
		block.Menu.Add(menuItem)
		block.Menu.ShowAll()
		return nil
	})
}

func (w *Window) updateBlock(update Update) error {
	block := w.findBlock(update.Name)
	if block == nil {
		return fmt.Errorf("couldn't find block %s", update.Name)
	}

	block.startUpdatingLabel()
	return nil
}

func (w *Window) removeBlock(remove Remove) error {
	block := w.findBlock(remove.Name)
	if block == nil {
		return fmt.Errorf("couldn't find block %s", remove.Name)
	}
	block.EventBox.Destroy()
	return nil
}

func (w *Window) findBlock(name string) *Block {
	for _, block := range w.blocks {
		if block.Name == name {
			return block
		}
	}
	return nil
}

func (w *Window) addBlockLeft(block *Block) {
	block.EventBox.SetHAlign(gtk.ALIGN_START)

	if w.lastLeftBlock != nil {
		w.gtkBar.AttachNextTo(block.EventBox, w.lastLeftBlock, gtk.POS_RIGHT, 1, 1)
	} else if w.lastCenterBlock != nil {
		w.gtkBar.AttachNextTo(block.EventBox, w.lastCenterBlock, gtk.POS_LEFT, 1, 1)
	} else if w.lastRightBlock != nil {
		w.gtkBar.AttachNextTo(block.EventBox, w.lastRightBlock, gtk.POS_LEFT, 1, 1)
	} else {
		w.gtkBar.Attach(block.EventBox, 0, 0, 1, 1)
	}
	w.lastLeftBlock = block.EventBox
}

func (w *Window) addBlockCenter(block *Block) {
	block.EventBox.SetHAlign(gtk.ALIGN_CENTER)
	block.EventBox.SetHExpand(true)
	block.Label.SetEllipsize(pango.ELLIPSIZE_END)

	if w.lastCenterBlock != nil {
		w.gtkBar.AttachNextTo(block.EventBox, w.lastCenterBlock, gtk.POS_RIGHT, 1, 1)
	} else if w.lastLeftBlock != nil {
		w.gtkBar.AttachNextTo(block.EventBox, w.lastLeftBlock, gtk.POS_RIGHT, 1, 1)
	} else if w.lastRightBlock != nil {
		w.gtkBar.AttachNextTo(block.EventBox, w.lastRightBlock, gtk.POS_LEFT, 1, 1)
	} else {
		w.gtkBar.Attach(block.EventBox, 0, 0, 1, 1)
	}
	w.lastCenterBlock = block.EventBox

}

func (w *Window) addBlockRight(block *Block) {
	block.EventBox.SetHAlign(gtk.ALIGN_END)

	if w.lastRightBlock != nil {
		w.gtkBar.AttachNextTo(block.EventBox, w.lastRightBlock, gtk.POS_RIGHT, 1, 1)
	} else if w.lastCenterBlock != nil {
		w.gtkBar.AttachNextTo(block.EventBox, w.lastCenterBlock, gtk.POS_RIGHT, 1, 1)
	} else if w.lastLeftBlock != nil {
		w.gtkBar.AttachNextTo(block.EventBox, w.lastLeftBlock, gtk.POS_RIGHT, 1, 1)
	} else {
		w.gtkBar.Attach(block.EventBox, 0, 0, 1, 1)
	}
	w.lastRightBlock = block.EventBox
}


