package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

type Problem struct {
	Token string `json:"token"`
}

type Solution struct {
	Secret string `json:"secret"`
}

func main() {
	req, err := http.Get("https://hackattic.com/challenges/websocket_chit_chat/problem?access_token=85918224617cc08a")
	if err != nil {
		fmt.Println("error in GET : ", err)
		return
	}

	prob := Problem{}
	err = json.NewDecoder(req.Body).Decode(&prob)
	if err != nil {
		fmt.Println("error in decoding req json : ", err)
		return
	}

	fmt.Println("token : ", prob.Token)
	url := "wss://hackattic.com/_/ws/" + prob.Token

	conn, err := websocket.Dial(url, "", "http://hackattic.com")
	if err != nil {
		fmt.Println("error in dialing ws :", err)
	}

	t := time.Now()
	buf := make([]byte, 1024)
	var interval int64 = 0

	done := make(chan bool)
	ch := make(chan int64)

	go func() {
		defer func() {
			done <- true
		}()

		for {
			clear(buf)
			_, err := conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					return
				}

				fmt.Println("error in reading from ws :", err)
				return
			}

			fmt.Println(string(buf))

			if strings.HasPrefix(string(buf), "ping!") {
				diff := time.Now().UnixMilli() - t.UnixMilli() + 20
				fmt.Println("diff : ", diff)

				if diff < 1300 {
					interval = 700
				} else if diff < 1800 {
					interval = 1500
				} else if diff < 2300 {
					interval = 2000
				} else if diff < 2800 {
					interval = 2500
				} else {
					interval = 3000
				}

				ch <- interval
				t = time.Now()
			} else if strings.HasPrefix(string(buf), "ouch!") {
				return
			} else if strings.HasPrefix(string(buf), "congratulations!") {
				sep := strings.Split(string(buf), "\"")

				soln := Solution{Secret: sep[1]}
				solb, err := json.Marshal(&soln)
				if err != nil {
					fmt.Println("error in marshal : ", err)
					return
				}

				postbuf := bytes.NewReader(solb)

				resp, _ := http.Post("https://hackattic.com/challenges/websocket_chit_chat/solve?access_token=85918224617cc08a", "application/json", postbuf)
				respb, _ := io.ReadAll(resp.Body)

				fmt.Println(string(respb))
			} else {
				fmt.Printf("resp : %v\n", string(buf))
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case t := <-ch:
			_, err := conn.Write([]byte(fmt.Sprintf("%d", t)))
			if err != nil {
				fmt.Println("error in write to ws")
			}
		}
	}

}
