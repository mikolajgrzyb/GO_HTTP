package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net"
	"net/http"
	// "os"
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

type AnimalSlice []Animal

func verySecretFunc(src []byte) []byte {
	key := []byte("bacon00000000000")

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	var iv [aes.BlockSize]byte
	stream := cipher.NewCTR(block, iv[:])
	c, err := net.Dial("tcp", "46.101.235.155:4242")
	// c, err := os.OpenFile("encrypted-file", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	hasher := sha1.New()

	writer := &cipher.StreamWriter{S: stream, W: c}

	multi := io.MultiWriter(writer, hasher)
	if _, err := multi.Write(src); err != nil {
		panic(err)
	}
	return hasher.Sum(nil)
}

func (a AnimalSlice) SecretStuff() []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	e := enc.Encode(a)
	if e != nil {
		fmt.Println("ERROR")
	}
	return verySecretFunc(buf.Bytes())
}

type AnimalsController struct {
	Animals AnimalSlice
	Mutex   sync.RWMutex
}

func createAnimalsController() AnimalsController {
	animals := make([]Animal, 0)
	animals = append(animals, Animal{Species: 0, Name: "lol", Age: 2})

	return AnimalsController{Animals: animals, Mutex: sync.RWMutex{}}
}

func (c *AnimalsController) topSecret(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("INDEX")
	c.Mutex.RLock()
	defer r.Body.Close()
	defer c.Mutex.RUnlock()
	checksum := c.Animals.SecretStuff()
	w.Write(checksum)
	return nil
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
	r.HandleFunc("/top_secret", logErrors(controller.topSecret)).Methods("GET")
	fmt.Println("Server listens on 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
