package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	escpos "go-escpos/utils"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/samber/lo"
)

type jokeAPI struct {
	Text     string
	Language string
}

type jokeJSON struct {
	Jokes []string `json:"jokes"`
}

const charsPerLine int = 44

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("First argument should backup joke json file path")
		os.Exit(0)
	}

	toggleLed(true, false)

	for {
		cmd := exec.Command("/bin/get-input", "1")
		stdout, _ := cmd.Output()
		if !strings.Contains(string(stdout), "HIGH") {
			continue
		}

		stopBlinking := make(chan bool)
		go blinkLed(stopBlinking)

		joke := ""

		transport := http.Transport{
			Dial: dialTimeout,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		client := http.Client{
			Transport: &transport,
		}

		response, err := client.Get("https://witzapi.de/api/joke")
		if err != nil {
			//use local jokes file if network error occured
			jsonFile, err := os.Open(os.Args[1])
			if err != nil {
				joke = "Lieber Benutzer,\n irgendwie komme ich nicht an die interne Witze-Datei.\n MfG Dein ECR"
			} else {
				defer jsonFile.Close()

				byteValue, _ := ioutil.ReadAll(jsonFile)
				var jjokes jokeJSON
				json.Unmarshal(byteValue, &jjokes)
				rand.Seed(time.Now().UnixNano())
				if len(jjokes.Jokes) == 0 {
					joke = "Lieber Benutzer,\n leider habe ich weder eine Internetverbindung, noch komme ich an die interne Witze-Datei.\n MfG Dein ECR"
				} else {
					joke = jjokes.Jokes[rand.Intn(len(jjokes.Jokes))]
				}
			}
		} else {
			responseData, _ := ioutil.ReadAll(response.Body)
			var jokes []jokeAPI
			json.Unmarshal(responseData, &jokes)
			joke = jokes[0].Text
		}

		fmt.Printf("Printing joke: %v", joke)

		socket, err := net.Dial("tcp", "192.168.1.1:232")
		if err != nil {
			println(err.Error())
		}

		w := bufio.NewWriter(socket)
		p := escpos.New(w)

		p.Init()
		p.WriteCP858(formatText(joke, charsPerLine))
		p.FormfeedD(2)

		p.Cut()

		w.Flush()
		socket.Close()
		stopBlinking <- true
	}
}

func formatText(text string, charsPerLine int) string {
	result := ""
	currentLine := ""
	for _, word := range strings.Split(text, " ") {
		if (len(word) + len(currentLine)) > charsPerLine {
			result += currentLine + "\n"
			currentLine = ""
		}
		currentLine += word + " "
	}
	result += currentLine
	return result
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, time.Duration(3*time.Second))
}

func toggleLed(state bool, wait bool) {
	cmd := exec.Command("/bin/set-output", "-o", "3.1", "-s", lo.Ternary(state, "close", "open"))
	if wait {
		cmd.Run()
	} else {
		cmd.Start()
	}
}

func blinkLed(stopBlinking chan bool) {
	for {
		select {
		case <-stopBlinking:
			toggleLed(true, false)
			return
		default:
			toggleLed(false, true)
			toggleLed(true, true)
		}
	}
}
