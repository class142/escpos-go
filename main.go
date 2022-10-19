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

	socket, err := net.Dial("tcp", "192.168.202.160:232")
	/* lizard := "" +
	"                       )/_       " +
	"             _.--..---''-,--c_    " +
	"        \\L..'           ._O__)_  " +
	",-.     _.+  _  \\..--( /         " +
	"  `\\.-''__.-' \\ (     \\_         " +
	"    `'''       `\\__   /\\         " +
	"                ')                "*/

	sudoku := "" +
		"╔═══╤═══╤═══╦═══╤═══╤═══╦═══╤═══╤═══╗\n" +
		"║ 8 │ 5 │   ║   │   │ 2 ║ 4 │   │   ║\n" +
		"╟───┼───┼───╫───┼───┼───╫───┼───┼───╢\n" +
		"║ 7 │ 2 │   ║   │   │   ║   │   │ 9 ║\n" +
		"╟───┼───┼───╫───┼───┼───╫───┼───┼───╢\n" +
		"║   │   │ 4 ║   │   │   ║   │   │   ║\n" +
		"╠═══╪═══╪═══╬═══╪═══╪═══╬═══╪═══╪═══╣\n" +
		"║   │   │   ║ 1 │   │ 7 ║   │   │ 2 ║\n" +
		"╟───┼───┼───╫───┼───┼───╫───┼───┼───╢\n" +
		"║ 3 │   │ 5 ║   │   │   ║ 9 │   │   ║\n" +
		"╟───┼───┼───╫───┼───┼───╫───┼───┼───╢\n" +
		"║   │ 4 │   ║   │   │   ║   │   │   ║\n" +
		"╠═══╪═══╪═══╬═══╪═══╪═══╬═══╪═══╪═══╣\n" +
		"║   │   │   ║   │ 8 │   ║   │ 7 │   ║\n" +
		"╟───┼───┼───╫───┼───┼───╫───┼───┼───╢\n" +
		"║   │ 1 │ 7 ║   │   │   ║   │   │   ║\n" +
		"╟───┼───┼───╫───┼───┼───╫───┼───┼───╢\n" +
		"║   │   │   ║   │ 3 │ 6 ║   │ 4 │   ║\n" +
		"╚═══╧═══╧═══╩═══╧═══╧═══╩═══╧═══╧═══╝\n"
	if err != nil {
		println(err.Error())
	}

	w1 := bufio.NewWriter(socket)
	p1 := escpos.New(w1)

	p1.Init()
	p1.WriteCP858(sudoku)
	p1.FormfeedD(2)

	p1.Cut()

	w1.Flush()

	return

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

	w := bufio.NewWriter(stream)
	p := escpos.New(w)

	if err != nil {
		println(err.Error())
	}

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
