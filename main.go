package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sync"

	"github.com/cep21/xdgbasedir"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("vbar", "A bar.")

	port = app.Flag("port", "Port to use for the command server.").Default("5643").OverrideDefaultFromEnvar("PORT").Int()

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
		log.Panic(err)
	}
	window = w

	go listenForCommands()

	go func() {
		err = executeConfig()
		if err != nil {
			log.Panic(err)
		}
	}()

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

func mutexLockedHandler(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		h.ServeHTTP(w, r)
	})
}

func listenForCommands() {
	http.HandleFunc("/add-css", mutexLockedHandler(addCSSHandler))
	http.HandleFunc("/add-block", mutexLockedHandler(addBlockHandler))
	http.HandleFunc("/add-menu", mutexLockedHandler(addMenuHandler))
	http.HandleFunc("/update", mutexLockedHandler(updateHandler))
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Panic(err)
	}
}

func sendAddCSS() {
	addCSS := AddCSS{
		Class: *flagAddCSSClass,
		Value: *flagAddCSSValue,
	}

	jsonValue, err := json.Marshal(addCSS)
	if err != nil {
		log.Panic(err)
	}

	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/add-css", *port),
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		log.Panic(err)
	}
	processResponse(resp.Body)
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
		log.Panic(err)
	}

	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/add-block", *port),
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		log.Panic(err)
	}
	processResponse(resp.Body)
}

func sendAddMenu() {
	addMenu := AddMenu{
		Name:    *flagAddMenuBlockName,
		Text:    *flagAddMenuText,
		Command: *flagAddMenuCommand,
	}
	jsonValue, err := json.Marshal(addMenu)
	if err != nil {
		log.Panic(err)
	}

	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/add-menu", *port),
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		log.Panic(err)
	}
	processResponse(resp.Body)
}

func sendUpdate() {
	update := Update{
		Name: *flagUpdateBlockName,
	}

	jsonValue, err := json.Marshal(update)
	if err != nil {
		log.Panic(err)
	}

	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/update", *port),
		"application/json",
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		log.Panic(err)
	}
	processResponse(resp.Body)
}

type ServerResponse struct {
	Error string
}

func processResponse(body io.ReadCloser) {
	defer body.Close()

	decoder := json.NewDecoder(body)
	var serverResponse ServerResponse
	err := decoder.Decode(&serverResponse)
	if err != nil {
		log.Fatal(err)
	}
	if serverResponse.Error != "" {
		log.Fatal(errors.New(serverResponse.Error))
	}
}

func writeError(w http.ResponseWriter, err error) {
	serverResponse := ServerResponse{Error: err.Error()}

	w.Header().Set("Content-Type", "application/json")
	result, err := json.Marshal(serverResponse)
	if err != nil {
		log.Panic(err)
	}
	io.WriteString(w, string(result))
}

func writeSuccess(w http.ResponseWriter) {
	serverResponse := ServerResponse{}

	w.Header().Set("Content-Type", "application/json")
	result, err := json.Marshal(serverResponse)
	if err != nil {
		log.Panic(err)
	}
	io.WriteString(w, string(result))
}

func addCSSHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var addCSS AddCSS
	err := decoder.Decode(&addCSS)
	if err != nil {
		writeError(w, err)
		return
	}
	defer r.Body.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	_, err = glib.IdleAdd(func() {
		defer wg.Done()
		err := window.applyCSS(addCSS)
		if err != nil {
			writeError(w, err)
		}
	})
	if err != nil {
		writeError(w, err)
		return
	}
	wg.Wait()
	writeSuccess(w)
}

func addBlockHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var block Block
	err := decoder.Decode(&block)
	if err != nil {
		writeError(w, err)
		return
	}
	defer r.Body.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	_, err = glib.IdleAdd(func() {
		defer wg.Done()
		err = window.addBlock(&block)
		if err != nil {
			writeError(w, err)
			return
		}
	})
	if err != nil {
		writeError(w, err)
		return
	}
	wg.Wait()
	writeSuccess(w)
}

func addMenuHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var addMenu AddMenu
	err := decoder.Decode(&addMenu)
	if err != nil {
		writeError(w, err)
		return
	}
	defer r.Body.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	_, err = glib.IdleAdd(func() {
		defer wg.Done()
		err = window.addMenu(addMenu)
		if err != nil {
			writeError(w, err)
			return
		}
	})
	if err != nil {
		writeError(w, err)
		return
	}
	wg.Wait()
	writeSuccess(w)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var update Update
	err := decoder.Decode(&update)
	if err != nil {
		writeError(w, err)
		return
	}
	defer r.Body.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	_, err = glib.IdleAdd(func() {
		defer wg.Done()
		err := window.updateBlock(update)
		if err != nil {
			writeError(w, err)
			return
		}
	})
	if err != nil {
		writeError(w, err)
		return
	}
	wg.Wait()
	writeSuccess(w)
}
