package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	escpos "go-escpos/utils"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	jokeapi "go-escpos/utils"

	"github.com/samber/lo"
	"github.com/tarm/serial"
)

/* type jokeAPI struct {
	Text     string
	Language string
}*/

type jokeJSON struct {
	Jokes []string `json:"jokes"`
}

const charsPerLine int = 44

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("First argument should backup joke json file path")
		os.Exit(0)
	}

	stopBlinking := make(chan bool)
	go blinkLed(stopBlinking)

	config := &serial.Config{
		Name:        "/devices/2_serial1",
		Baud:        9600,
		ReadTimeout: 1,
		Size:        8,
	}

	stream, err := serial.OpenPort(config)

	if err != nil {
		println(err.Error())
	}

	w := bufio.NewWriter(stream)
	p := escpos.New(w)

	api := jokeapi.Init()
	api.Set(jokeapi.Params{JokeType: "single"})

	stopBlinking <- true

	for {
		cmd := exec.Command("/bin/get-input", "1")
		stdout, _ := cmd.Output()
		if !strings.Contains(string(stdout), "HIGH") {
			continue
		}

		stopBlinking := make(chan bool)
		go blinkLed(stopBlinking)

		joke := ""

		/* transport := http.Transport{
			Dial: dialTimeout,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		client := http.Client{
			Transport: &transport,
		}

		response, err := client.Get("https://witzapi.de/api/joke") */

		response, err := api.Fetch()

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
			joke = response.Joke[0]
		}

		fmt.Printf("Printing joke: %v \n", joke)

		p.Init()
		p.WriteCP858(formatText(joke, charsPerLine))
		p.FormfeedD(2)

		p.Cut()

		w.Flush()
		//stream.Close()
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
