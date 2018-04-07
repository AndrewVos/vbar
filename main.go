package main

/*
#cgo pkg-config: gdk-3.0
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
#include <gdk/gdk.h>

static GtkWindow * toGtkWindow(void *p)
{
  return (GTK_WINDOW(p));
}

static GtkWidget * toGtkWidget(void *p)
{
  return (GTK_WIDGET(p));
}

static GtkMenu * toGtkMenu(void *p)
{
  return (GTK_MENU(p));
}

static GdkDisplay * toGdkDisplay(void *p)
{
	return (GDK_DISPLAY(p));
}

void set_strut_properties(GtkWindow *window,
				long left, long right, long top, long bottom,
 				long left_start_y, long left_end_y,
 				long right_start_y, long right_end_y,
 				long top_start_x, long top_end_x,
 				long bottom_start_x, long bottom_end_x) {
	gulong data[12] = {0};
	data[0] = left; data[1] = right; data[2] = top; data[3] = bottom;
	data[4] = left_start_y; data[5] = left_end_y;
	data[6] = right_start_y; data[7] = right_end_y;
	data[8] = top_start_x; data[9] = top_end_x;
	data[10] = bottom_start_x; data[11] = bottom_end_x;

	gdk_property_change(gtk_widget_get_window(GTK_WIDGET(window)),
				gdk_atom_intern("_NET_WM_STRUT_PARTIAL", FALSE),
				gdk_atom_intern ("CARDINAL", FALSE),
				32, GDK_PROP_MODE_REPLACE, (unsigned char *)data, 12);

	gdk_property_change(gtk_widget_get_window(GTK_WIDGET(window)),
				gdk_atom_intern("_NET_WM_STRUT", FALSE),
				gdk_atom_intern("CARDINAL", FALSE), 32, GDK_PROP_MODE_REPLACE,
				(unsigned char *) &data, 4);

}
*/
import "C"

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	"unsafe"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("vbar", "A bar.")

	commandStart = app.Command("start", "Start vbar.")

	commandAddCSS   = app.Command("add-css", "Add CSS.")
	flagAddCSSClass = commandAddCSS.Flag("class", "CSS Class name.").Required().String()
	flagAddCSSValue = commandAddCSS.Flag("css", "CSS value.").Required().String()

	commandAddBlock         = app.Command("add-block", "Add a new block.")
	flagAddBlockName        = commandAddBlock.Flag("name", "Block name.").Required().String()
	flagAddBlockLeft        = commandAddBlock.Flag("left", "Add block to the left.").Bool()
	flagAddBlockCenter      = commandAddBlock.Flag("center", "Add block to the center.").Bool()
	flagAddBlockRight       = commandAddBlock.Flag("right", "Add block to the right.").Bool()
	flagAddBlockText        = commandAddBlock.Flag("text", "Block text.").String()
	flagAddBlockCommand     = commandAddBlock.Flag("command", "Command to execute.").String()
	flagAddBlockTailCommand = commandAddBlock.Flag("tail-command", "Command to tail.").String()
	flagAddBlockInterval    = commandAddBlock.Flag("interval", "Interval in seconds to execute command.").Int()

	commandAddMenu       = app.Command("add-menu", "Add a menu to a block.")
	flagAddMenuBlockName = commandAddMenu.Flag("name", "Block name.").Required().String()
	flagAddMenuText      = commandAddMenu.Flag("text", "Menu text.").Required().String()
	flagAddMenuCommand   = commandAddMenu.Flag("command", "Command to execute when activating the menu.").Required().String()

	commandUpdate       = app.Command("update", "Trigger a block update.")
	flagUpdateBlockName = commandUpdate.Flag("name", "Block name.").Required().String()

	window *gtk.Window
	panel  *gtk.Grid
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case commandStart.FullCommand():
		startVbar()
	case commandAddCSS.FullCommand():
		sendAddCSS()
	case commandAddBlock.FullCommand():
		sendAddBlock()
	case commandAddMenu.FullCommand():
		sendAddMenu()
	case commandUpdate.FullCommand():
		sendUpdate()
	}
}

type blockOptions struct {
	EventBox    *gtk.EventBox
	Label       *gtk.Label
	Menu        *gtk.Menu
	Name        string
	Text        string
	Left        bool
	Center      bool
	Right       bool
	Command     string
	TailCommand string
	Interval    int
}

type updateOptions struct {
	Name string
}

func (bo blockOptions) updateLabel() {
	cmd := exec.Command("/bin/bash", "-c", bo.Command)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.Output()
	if err == nil {
		bo.Label.SetText(strings.TrimSpace(string(stdout)))
	} else {
		log.Printf("Command finished with error: %v", err)
		bo.Label.SetText("ERROR")
	}
}

func (bo blockOptions) updateLabelForever() {
	go func() {
		cmd := exec.Command("/bin/bash", "-c", bo.TailCommand)
		cmd.Stderr = os.Stderr

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("Couldn't get a stdout from command: %v", err)
			bo.Label.SetText("ERROR")
			return
		}
		err = cmd.Start()
		if err != nil {
			log.Printf("Command finished with error: %v", err)
			bo.Label.SetText("ERROR")
			return
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			bo.Label.SetText(strings.TrimSpace(scanner.Text()))
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Couldn't read from command stdout: %v", err)
			bo.Label.SetText("ERROR")
			return
		}
	}()
}

func applyClass(widget *gtk.Widget, class string) {
	styleContext, err := widget.GetStyleContext()
	if err != nil {
		log.Fatal(err)
	}
	styleContext.AddClass(class)
}

// Rectangle is just a rectangle.
type Rectangle struct {
	X      int
	Y      int
	Width  int
	Height int
}

func enableTransparency(window *gtk.Window) error {
	screen, err := window.GetScreen()
	if err != nil {
		return err
	}

	visual, err := screen.GetRGBAVisual()
	if err != nil {
		return err
	}

	if visual != nil && screen.IsComposited() {
		window.SetVisual(visual)
	}

	return nil
}

func getMonitorDimensions(window *gtk.Window) (Rectangle, error) {
	screen, err := window.GetScreen()
	if err != nil {
		return Rectangle{}, err
	}
	display, err := screen.GetDisplay()
	if err != nil {
		return Rectangle{}, err
	}

	geometry := C.GdkRectangle{}
	displayPointer := unsafe.Pointer(display.GObject)
	gdkDisplay := C.toGdkDisplay(displayPointer)
	monitor := C.gdk_display_get_primary_monitor(gdkDisplay)
	C.gdk_monitor_get_geometry(monitor, &geometry)
	return Rectangle{
		X:      int(geometry.x),
		Y:      int(geometry.y),
		Width:  int(geometry.width),
		Height: int(geometry.height),
	}, nil
}

var blocks []*blockOptions

func buildEventBox(options *blockOptions) {
	blocks = append(blocks, options)

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
		addBlockLeft(eventBox)
	} else if options.Center {
		addBlockCenter(eventBox)
	} else if options.Right {
		addBlockRight(eventBox)
	}

	//TODO: click_command

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
}

type serverResult struct {
	Success bool
}

type cssOptions struct {
	Class string
	Value string
}

func sendAddCSS() {
	options := cssOptions{
		Class: *flagAddCSSClass,
		Value: *flagAddCSSValue,
	}

	jsonValue, err := json.Marshal(options)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(
		"http://localhost:5643/add-css",
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result serverResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatal(err)
	}
	if result.Success == false {
		log.Fatal("Command failed.")
	}
}

func sendAddBlock() {
	options := blockOptions{
		Name:        *flagAddBlockName,
		Text:        *flagAddBlockText,
		Left:        *flagAddBlockLeft,
		Center:      *flagAddBlockCenter,
		Right:       *flagAddBlockRight,
		Command:     *flagAddBlockCommand,
		TailCommand: *flagAddBlockTailCommand,
		Interval:    *flagAddBlockInterval,
	}
	jsonValue, err := json.Marshal(options)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(
		"http://localhost:5643/add-block",
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result serverResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatal(err)
	}
	if result.Success == false {
		log.Fatal("Command failed.")
	}
}

type menuOptions struct {
	Name    string
	Text    string
	Command string
}

func sendAddMenu() {
	options := menuOptions{
		Name:    *flagAddMenuBlockName,
		Text:    *flagAddMenuText,
		Command: *flagAddMenuCommand,
	}
	jsonValue, err := json.Marshal(options)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(
		"http://localhost:5643/add-menu",
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result serverResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatal(err)
	}
	if result.Success == false {
		log.Fatal("Command failed.")
	}
}

func sendUpdate() {
	options := updateOptions{
		Name: *flagUpdateBlockName,
	}

	jsonValue, err := json.Marshal(options)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(
		"http://localhost:5643/update",
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result serverResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatal(err)
	}
	if result.Success == false {
		log.Fatal("Command failed.")
	}
}

func addBlockHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var options blockOptions
	err := decoder.Decode(&options)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	buildEventBox(&options)

	window.ShowAll()

	result := serverResult{Success: true}
	jsonValue, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, string(jsonValue))
}

func addMenuHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var options menuOptions
	err := decoder.Decode(&options)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	for _, block := range blocks {
		if block.Name == options.Name {
			if block.Menu == nil {
				menu, err := gtk.MenuNew()
				if err != nil {
					log.Fatal(err)
				}

				block.Menu = menu
				applyClass(&block.Menu.Widget, "menu")

				block.EventBox.Connect("button-release-event", func() {
					menuPointer := unsafe.Pointer(menu.GObject)
					gtkMenu := C.toGtkMenu(menuPointer)

					widgetPointer := unsafe.Pointer(block.EventBox.Widget.GObject)
					gtkWidget := C.toGtkWidget(widgetPointer)

					C.gtk_menu_popup_at_widget(
						gtkMenu,
						gtkWidget,
						C.GDK_GRAVITY_SOUTH_WEST,
						C.GDK_GRAVITY_NORTH_WEST,
						nil,
					)
				})

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
				menu.Add(menuItem)
				menu.ShowAll()

			}
		}
	}

	result := serverResult{Success: true}
	jsonValue, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, string(jsonValue))
}

func addCSSHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var options cssOptions
	err := decoder.Decode(&options)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	cssAdder.Add(options)

	result := serverResult{Success: true}
	jsonValue, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, string(jsonValue))
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var options updateOptions
	err := decoder.Decode(&options)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	for _, block := range blocks {
		if block.Name == options.Name {
			block.updateLabel()
			break
		}
	}

	result := serverResult{Success: true}
	jsonValue, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, string(jsonValue))
}

var cssAdder CSSAdder

// CSSAdder applies CSS to the bar.
type CSSAdder struct {
	Screen     *gdk.Screen
	cssOptions []cssOptions
	provider   *gtk.CssProvider
}

// Add applies CSS to the bar.
func (ca *CSSAdder) Add(options cssOptions) {
	ca.cssOptions = append(ca.cssOptions, options)

	if ca.provider == nil {
		provider, err := gtk.CssProviderNew()
		if err != nil {
			log.Fatal(err)
		}
		ca.provider = provider
		gtk.AddProviderForScreen(ca.Screen, provider, 0)
	}

	css := ""
	for _, options := range ca.cssOptions {
		css += fmt.Sprintf(".%s { %s }\n", options.Class, options.Value)
	}
	err := ca.provider.LoadFromData(css)
	if err != nil {
		log.Fatal(err)
	}
}

func updateDimensions() error {
	window.ShowAll()

	monitorDimensions, err := getMonitorDimensions(window)
	if err != nil {
		return err
	}

	window.SetSizeRequest(monitorDimensions.Width, -1)

	p := unsafe.Pointer(window.GObject)
	w := C.toGtkWindow(p)

	C.set_strut_properties(
		w,
		0, 0, C.long(panel.GetAllocatedHeight()), 0, /* strut-left, strut-right, strut-top, strut-bottom */
		0, 0, /* strut-left-start-y, strut-left-end-y */
		0, 0, /* strut-right-start-y, strut-right-end-y */
		C.long(monitorDimensions.X), C.long(monitorDimensions.X+monitorDimensions.Width-1), /* strut-top-start-x, strut-top-end-x */
		0, 0, /* strut-bottom-start-x, strut-bottom-end-x */
	)
	return nil
}

func startVbar() {
	gtk.Init(nil)

	err := errors.New("Hi")
	window, err = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	window.SetAppPaintable(true)
	window.SetDecorated(false)
	window.SetResizable(false)
	window.SetSkipPagerHint(true)
	window.SetSkipTaskbarHint(true)
	window.SetTypeHint(gdk.WINDOW_TYPE_HINT_DOCK)
	window.SetVExpand(false)
	window.SetPosition(gtk.WIN_POS_NONE)
	window.Move(0, 0)
	window.SetSizeRequest(-1, -1)

	window.Connect("destroy", func() {
		gtk.MainQuit()
	})

	window.Connect("realize", func() {
		updateDimensions()
	})

	screen, err := window.GetScreen()
	if err != nil {
		log.Fatal(err)
	}
	cssAdder = CSSAdder{
		Screen: screen,
	}

	panel, err = gtk.GridNew()
	if err != nil {
		log.Fatal(err)
	}
	applyClass(&panel.Widget, "panel")
	window.Add(panel)

	enableTransparency(window)

	go listen()

	cmd := exec.Command("/bin/bash", "-c", "/home/andrewvos/.config/vbar/vbarrc")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
	}

	gtk.Main()

}

func listen() {
	http.HandleFunc("/add-block", addBlockHandler)
	http.HandleFunc("/add-menu", addMenuHandler)
	http.HandleFunc("/add-css", addCSSHandler)
	http.HandleFunc("/update", updateHandler)
	err := http.ListenAndServe(":5643", nil)
	if err != nil {
		log.Fatal(err)
	}
}

var lastLeftBlock *gtk.EventBox
var lastCenterBlock *gtk.EventBox
var lastRightBlock *gtk.EventBox

func addBlockLeft(block *gtk.EventBox) {
	block.SetHAlign(gtk.ALIGN_START)

	if lastLeftBlock != nil {
		panel.AttachNextTo(block, lastLeftBlock, gtk.POS_RIGHT, 1, 1)
	} else if lastCenterBlock != nil {
		panel.AttachNextTo(block, lastCenterBlock, gtk.POS_LEFT, 1, 1)
	} else if lastRightBlock != nil {
		panel.AttachNextTo(block, lastRightBlock, gtk.POS_LEFT, 1, 1)
	} else {
		panel.Attach(block, 0, 0, 1, 1)
	}
	lastLeftBlock = block
}

func addBlockCenter(block *gtk.EventBox) {
	block.SetHAlign(gtk.ALIGN_CENTER)
	block.SetHExpand(true)

	if lastCenterBlock != nil {
		panel.AttachNextTo(block, lastCenterBlock, gtk.POS_RIGHT, 1, 1)
	} else if lastLeftBlock != nil {
		panel.AttachNextTo(block, lastLeftBlock, gtk.POS_RIGHT, 1, 1)
	} else if lastRightBlock != nil {
		panel.AttachNextTo(block, lastRightBlock, gtk.POS_LEFT, 1, 1)
	} else {
		panel.Attach(block, 0, 0, 1, 1)
	}
	lastCenterBlock = block

}

func addBlockRight(block *gtk.EventBox) {
	block.SetHAlign(gtk.ALIGN_END)

	if lastRightBlock != nil {
		panel.AttachNextTo(block, lastRightBlock, gtk.POS_RIGHT, 1, 1)
	} else if lastCenterBlock != nil {
		panel.AttachNextTo(block, lastCenterBlock, gtk.POS_RIGHT, 1, 1)
	} else if lastLeftBlock != nil {
		panel.AttachNextTo(block, lastLeftBlock, gtk.POS_RIGHT, 1, 1)
	} else {
		panel.Attach(block, 0, 0, 1, 1)
	}
	lastRightBlock = block
}
