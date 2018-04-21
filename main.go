package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

type CommandFunc func(body []byte) error

func listenForCommands() {
	handler := func(c CommandFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				writeError(w, err)
				return
			}
			defer r.Body.Close()

			mutex.Lock()
			defer mutex.Unlock()
			err = c(body)
			if err != nil {
				writeError(w, err)
				return
			}
			writeSuccess(w)
		})
	}

	http.HandleFunc("/add-css", handler(func(body []byte) error {
		var wg sync.WaitGroup
		wg.Add(1)

		var command AddCSS
		err := json.Unmarshal(body, &command)
		if err != nil {
			return err
		}

		var commandError error
		_, err = glib.IdleAdd(func() {
			defer wg.Done()
			commandError = window.applyCSS(command)
		})
		if err != nil {
			return err
		}
		wg.Wait()
		return commandError
	}))

	http.HandleFunc("/add-block", handler(func(body []byte) error {
		var wg sync.WaitGroup
		wg.Add(1)

		var command *Block
		err := json.Unmarshal(body, &command)
		if err != nil {
			return err
		}

		var commandError error
		_, err = glib.IdleAdd(func() {
			defer wg.Done()
			commandError = window.addBlock(command)
		})
		if err != nil {
			return err
		}
		wg.Wait()
		return commandError
	}))

	http.HandleFunc("/add-menu", handler(func(body []byte) error {
		var wg sync.WaitGroup
		wg.Add(1)

		var command AddMenu
		err := json.Unmarshal(body, &command)
		if err != nil {
			return err
		}

		var commandError error
		_, err = glib.IdleAdd(func() {
			defer wg.Done()
			commandError = window.addMenu(command)
		})
		if err != nil {
			return err
		}
		wg.Wait()
		return commandError
	}))

	http.HandleFunc("/update", handler(func(body []byte) error {
		var wg sync.WaitGroup
		wg.Add(1)

		var command Update
		err := json.Unmarshal(body, &command)
		if err != nil {
			return err
		}

		var commandError error
		_, err = glib.IdleAdd(func() {
			defer wg.Done()
			commandError = window.updateBlock(command)
		})
		if err != nil {
			return err
		}
		wg.Wait()
		return commandError
	}))

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
