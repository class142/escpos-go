package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
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

	for {
		cmd := exec.Command("/bin/get-input", "1")
		stdout, _ := cmd.Output()
		if !strings.Contains(string(stdout), "HIGH") {
			continue
		}

		socket, err := net.Dial("tcp", "192.168.1.1:232")
		if err != nil {
			println(err.Error())
		}

		w := bufio.NewWriter(socket)
		p := escpos.New(w)
		joke := ""
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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
		socket.Close()
	}
}
