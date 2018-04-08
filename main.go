package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/cep21/xdgbasedir"
	"github.com/gotk3/gotk3/gtk"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("vbar", "A bar.")

	commandStart = app.Command("start", "Start vbar.")

	commandAddCSS   = app.Command("add-css", "Add CSS.")
	flagAddCSSClass = commandAddCSS.Flag("class", "CSS Class name.").Required().String()
	flagAddCSSValue = commandAddCSS.Flag("css", "CSS value.").Required().String()

	commandAddBlock          = app.Command("add-block", "Add a new block.")
	flagAddBlockName         = commandAddBlock.Flag("name", "Block name.").Required().String()
	flagAddBlockLeft         = commandAddBlock.Flag("left", "Add block to the left.").Bool()
	flagAddBlockCenter       = commandAddBlock.Flag("center", "Add block to the center.").Bool()
	flagAddBlockRight        = commandAddBlock.Flag("right", "Add block to the right.").Bool()
	flagAddBlockText         = commandAddBlock.Flag("text", "Block text.").String()
	flagAddBlockCommand      = commandAddBlock.Flag("command", "Command to execute.").String()
	flagAddBlockTailCommand  = commandAddBlock.Flag("tail-command", "Command to tail.").String()
	flagAddBlockInterval     = commandAddBlock.Flag("interval", "Interval in seconds to execute command.").Int()
	flagAddBlockClickCommand = commandAddBlock.Flag("click-command", "Command to execute when clicking on the block.").String()

	commandAddMenu       = app.Command("add-menu", "Add a menu to a block.")
	flagAddMenuBlockName = commandAddMenu.Flag("name", "Block name.").Required().String()
	flagAddMenuText      = commandAddMenu.Flag("text", "Menu text.").Required().String()
	flagAddMenuCommand   = commandAddMenu.Flag("command", "Command to execute when activating the menu.").Required().String()

	commandUpdate       = app.Command("update", "Trigger a block update.")
	flagUpdateBlockName = commandUpdate.Flag("name", "Block name.").Required().String()

	window *Window
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
	result, err := decodeHandlerResult(resp)

	if result.Success == false {
		log.Fatal("Command failed.")
	}
}

func sendAddBlock() {
	options := blockOptions{
		Name:         *flagAddBlockName,
		Text:         *flagAddBlockText,
		Left:         *flagAddBlockLeft,
		Center:       *flagAddBlockCenter,
		Right:        *flagAddBlockRight,
		Command:      *flagAddBlockCommand,
		TailCommand:  *flagAddBlockTailCommand,
		Interval:     *flagAddBlockInterval,
		ClickCommand: *flagAddBlockClickCommand,
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

	result, err := decodeHandlerResult(resp)
	if err != nil {
		log.Fatal(err)
	} else if result.Success == false {
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

	result, err := decodeHandlerResult(resp)
	if err != nil {
		log.Fatal(err)
	} else if result.Success == false {
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

	result, err := decodeHandlerResult(resp)
	if err != nil {
		log.Fatal(err)
	} else if result.Success == false {
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

	err = window.addBlock(&options)
	if err != nil {
		fmt.Fprintf(w, dumpHandlerResult(false))
		return
	}

	fmt.Fprintf(w, dumpHandlerResult(true))
}

func addMenuHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var options menuOptions
	err := decoder.Decode(&options)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	err = window.addMenu(options)
	if err != nil {
		fmt.Fprintf(w, dumpHandlerResult(false))
		return
	}

	fmt.Fprintf(w, dumpHandlerResult(true))
}

type handlerResult struct {
	Success bool
}

func decodeHandlerResult(response *http.Response) (handlerResult, error) {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return handlerResult{}, err
	}

	var result handlerResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return handlerResult{}, err
	}

	return result, nil
}

func dumpHandlerResult(success bool) string {
	result := struct {
		Success bool
	}{Success: success}

	jsonValue, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	return string(jsonValue)
}

func addCSSHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var options cssOptions
	err := decoder.Decode(&options)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	window.applyCSS(options)

	fmt.Fprintf(w, dumpHandlerResult(true))
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var options updateOptions
	err := decoder.Decode(&options)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	window.updateBlock(options)

	fmt.Fprintf(w, dumpHandlerResult(true))
}

func startVbar() {
	gtk.Init(nil)

	w, err := WindowNew()
	if err != nil {
		log.Fatal(err)
	}
	window = w

	go listenForCommands()
	err = executeConfig()
	if err != nil {
		log.Println(err)
	}

	gtk.Main()
}

func executeConfig() error {
	configurationDirectory, err := xdgbasedir.ConfigHomeDirectory()
	if err != nil {
		return err
	}
	configurationFilePath := path.Join(configurationDirectory, "vbar", "vbarrc")

	cmd := exec.Command("/bin/bash", "-c", configurationFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func listenForCommands() {
	http.HandleFunc("/add-block", addBlockHandler)
	http.HandleFunc("/add-menu", addMenuHandler)
	http.HandleFunc("/add-css", addCSSHandler)
	http.HandleFunc("/update", updateHandler)
	err := http.ListenAndServe(":5643", nil)
	if err != nil {
		log.Fatal(err)
	}
}
