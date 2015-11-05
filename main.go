package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

const (
	chicken Species = iota
	pig
	cow
)

type Species int

// Animal is a farm's SKU.
type Animal struct {
	Species Species `json:"species,string"`
	Name    string  `json:"name"`
	Age     int     `json:"age,string"`
}

var animals = make([]Animal, 0)
var validPath = regexp.MustCompile("animals")

func animalsIndex(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(animals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)

	w.Write(js)
}

func animalsCreate(w http.ResponseWriter, r *http.Request) {
	var animal Animal
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&animal)
	if err != nil {
		panic(err)
	}
	animals = append(animals, animal)
	w.WriteHeader(200)
	js, err := json.Marshal(animals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(200)
	w.Write(js)
}

func AnimalsController(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("INDEX")
		animalsIndex(w, r)
	case "POST":
		fmt.Println("CREATE")
		animalsCreate(w, r)
	}
}

func setJsonHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	AnimalsController(w, r)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r)
	}
}

func main() {
	http.HandleFunc("/animals", makeHandler(setJsonHeader))
	http.ListenAndServe(":8080", nil)
}
