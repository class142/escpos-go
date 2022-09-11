package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	escpos "go-escpos/utils"
)

type jokeAPI struct {
	Text     string
	Language string
}

type jokeJSON struct {
	Jokes []string `json:"jokes"`
}

func main() {
	socket, err := net.Dial("tcp", "10.9.9.100:232")
	if err != nil {
		println(err.Error())
	}
	defer socket.Close()

	w := bufio.NewWriter(socket)
	p := escpos.New(w)
	joke := ""
	response, err := http.Get("https://witzapi.de/api/joke")
	if err != nil {
		//use local jokes file if network error occured
		jsonFile, err := os.Open("jokes.json")
		if err != nil {
			//continue
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)
		var jjokes jokeJSON
		json.Unmarshal(byteValue, &jjokes)
		rand.Seed(time.Now().UnixNano())
		joke = jjokes.Jokes[rand.Intn(len(jjokes.Jokes))]
	} else {
		responseData, _ := ioutil.ReadAll(response.Body)
		var jokes []jokeAPI
		json.Unmarshal(responseData, &jokes)
		joke = jokes[0].Text
	}

	p.Init()

	p.WriteCP858(joke)
	p.FormfeedD(2)

	p.Cut()

	w.Flush()
}
