package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sync"

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
	mutex  = &sync.Mutex{}
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case commandStart.FullCommand():
		launch()
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

func launch() {
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
	http.HandleFunc("/add-css", addCSSHandler)
	http.HandleFunc("/add-block", addBlockHandler)
	http.HandleFunc("/add-menu", addMenuHandler)
	http.HandleFunc("/update", updateHandler)
	err := http.ListenAndServe(":5643", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func sendAddCSS() {
	addCSS := AddCSS{
		Class: *flagAddCSSClass,
		Value: *flagAddCSSValue,
	}

	jsonValue, err := json.Marshal(addCSS)
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
	block := Block{
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
	jsonValue, err := json.Marshal(block)
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

func sendAddMenu() {
	addMenu := AddMenu{
		Name:    *flagAddMenuBlockName,
		Text:    *flagAddMenuText,
		Command: *flagAddMenuCommand,
	}
	jsonValue, err := json.Marshal(addMenu)
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
	update := Update{
		Name: *flagUpdateBlockName,
	}

	jsonValue, err := json.Marshal(update)
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

func addCSSHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	decoder := json.NewDecoder(r.Body)
	var addCSS AddCSS
	err := decoder.Decode(&addCSS)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	window.applyCSS(addCSS)

	fmt.Fprintf(w, dumpHandlerResult(true))
}

func addBlockHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	decoder := json.NewDecoder(r.Body)
	var block Block
	err := decoder.Decode(&block)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	err = window.addBlock(&block)
	if err != nil {
		fmt.Fprintf(w, dumpHandlerResult(false))
		return
	}

	fmt.Fprintf(w, dumpHandlerResult(true))
}

func addMenuHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	decoder := json.NewDecoder(r.Body)
	var addMenu AddMenu
	err := decoder.Decode(&addMenu)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	err = window.addMenu(addMenu)
	if err != nil {
		fmt.Fprintf(w, dumpHandlerResult(false))
		return
	}

	fmt.Fprintf(w, dumpHandlerResult(true))
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	decoder := json.NewDecoder(r.Body)
	var update Update
	err := decoder.Decode(&update)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	window.updateBlock(update)

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
