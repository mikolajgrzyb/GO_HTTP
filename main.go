package main

import (
	"encoding/json"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	chicken Species = iota
	pig
	cow
)

type Species int

// Animal is a farm's SKU.
type Animal struct {
	Species Species `json:"species,	string"`
	Name    string  `json:"name"valid:"required"`
	Age     int     `json:"age,string"valid:"required,int"`
}

type AnimalsController struct {
	Animals []Animal
	Mutex   sync.RWMutex
}

func createAnimalsController() AnimalsController {
	return AnimalsController{Animals: make([]Animal, 0), Mutex: sync.RWMutex{}}
}

func (c *AnimalsController) index(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("INDEX")
	c.Mutex.RLock()
	defer r.Body.Close()
	defer c.Mutex.RUnlock()
	enc := json.NewEncoder(w)
	return enc.Encode(c.Animals)
}

func (c *AnimalsController) create(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("CREATE")
	c.Mutex.Lock()
	defer r.Body.Close()
	defer c.Mutex.Unlock()
	var animal Animal
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&animal)
	if err != nil {
		return err
	}
	result, error := govalidator.ValidateStruct(animal)
	if error != nil {
		return error
	}
	fmt.Println(result)
	c.Animals = append(c.Animals, animal)
	return nil
}

func logErrors(fn func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		error := fn(w, r)
		if error != nil {
			w.WriteHeader(422)
			errorString := error.Error()
			errorArray := strings.Split(errorString, ";")
			var newErrorArray []map[string]string
			for _, errorString := range errorArray {
				splittedError := strings.Split(errorString, ": ")
				fmt.Println(splittedError[0])
				newErrorArray = append(newErrorArray, map[string]string{splittedError[0]: splittedError[1]})
			}
			fmt.Println(newErrorArray[0])
			dec := json.NewEncoder(w)
			dec.Encode(newErrorArray)
		}
	}
}

func main() {
	r := mux.NewRouter()
	controller := createAnimalsController()
	r.HandleFunc("/animals", logErrors(controller.index)).Methods("GET")
	r.HandleFunc("/animals", logErrors(controller.create)).Methods("POST")
	fmt.Println("Server listens on 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
