package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"
)

type Note struct {
	Title string
	Body  string
	Date  time.Time
}

type NoteSlice []Note

func (ns NoteSlice) ToJSON() string {
	b, err := json.Marshal(ns)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

var (
	datastoreLocation = flag.String("ds", "datastore.json", "where to read/write the datastore")
	indexTemplate     *template.Template
	notes             NoteSlice
)

func main() {
	flag.Parse()

	notes = readDatastore(*datastoreLocation)

	var err error
	indexTemplate, err = template.New("index.html").ParseFiles("templates/index.html")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/save", save)
	http.HandleFunc("/", index)
	http.ListenAndServe(":2222", nil)
}

func readDatastore(datastoreLocation string) NoteSlice {
	file := tryOpenFileReader(datastoreLocation)
	var notes NoteSlice
	if file == nil {
		notes = NoteSlice{}
	} else {
		decoder := json.NewDecoder(file)
		err := decoder.Decode(&notes)
		if err != nil {
			fmt.Println("failed to parse datastore, it may have been empty.")
			notes = NoteSlice{}
		}
	}
	return notes
}

func index(w http.ResponseWriter, r *http.Request) {
	indexTemplate.Execute(w, notes)
}

func save(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var tempNotes NoteSlice
	err := decoder.Decode(&tempNotes)
	if err != nil {
		// return err.Error()
		fmt.Fprintf(w, "failure")
		return
	}
	defer r.Body.Close()

	notes = tempNotes

	b, err := json.Marshal(notes)
	if err != nil {
		// return err.Error()
		fmt.Fprintf(w, "failure")
		return
	}

	file, err := os.Create(*datastoreLocation)
	if err != nil {
		// return err.Error()
		fmt.Fprintf(w, "failure")
		return
	}
	defer file.Close()

	n, err := file.Write(b)

	if n != len(b) {
		fmt.Fprintf(w, "failure")
		return
	}

	fmt.Fprintf(w, "success")
}

func openFileReaderOrFail(filename string) *os.File {
	file := tryOpenFileReader(filename)
	if file == nil {
		panic(fmt.Sprintf("Failed to open: '%s'\n", filename))
	}
	return file
}

func tryOpenFileReader(filename string) *os.File {
	file, _ := os.Open(filename)
	return file
}
