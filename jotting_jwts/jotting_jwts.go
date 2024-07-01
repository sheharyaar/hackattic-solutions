package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type Problem struct {
	JWTSecret string `json:"jwt_secret"`
}

type Solution struct {
	Solution string     `json:"solution"`
	Mu       sync.Mutex `json:"-"`
}

type PostURL struct {
	AppURL string `json:"app_url"`
}

var Secret string
var Soln Solution

func handleJWTs(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error in reading POST body : ", err)
		http.Error(w, "invalid body", 400)
		return
	}
	// check if jwt is valid
	str, err := CheckJWT(string(b), Secret)
	if err != nil {
		fmt.Println("check JWT failed : ", err)
		return
	}

	if str != "" {
		Soln.Mu.Lock()
		Soln.Solution += str
		Soln.Mu.Unlock()
	} else {
		// send the string
		solb, err := json.Marshal(&Soln)
		if err != nil {
			fmt.Println("error in json marshal solution : ", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(solb)
	}

	fmt.Printf("Str: %s\tstr: %s\n", Soln.Solution, str)
}

func main() {
	req, err := http.Get("https://hackattic.com/challenges/jotting_jwts/problem?access_token=85918224617cc08a")
	if err != nil {
		fmt.Println("error in GET : ", err)
		return
	}

	prob := Problem{}
	if err = json.NewDecoder(req.Body).Decode(&prob); err != nil {
		fmt.Println("error in decoding json : ", err)
		return
	}

	Secret = prob.JWTSecret

	go func() {
		time.Sleep(time.Second * 2)
		posturl := PostURL{AppURL: "https://07f7-49-37-67-15.ngrok-free.app"}
		b, err := json.Marshal(&posturl)
		if err != nil {
			fmt.Println("error in json marshal post : ", err)
			return
		}

		buf := bytes.NewReader(b)

		postb, err := http.Post("https://hackattic.com/challenges/jotting_jwts/solve?access_token=85918224617cc08a", "application/json", buf)
		if err != nil {
			fmt.Println("error in POST : ", err)
			return
		}

		post, _ := io.ReadAll(postb.Body)
		fmt.Println(string(post))
	}()

	// create server and serve requests
	http.HandleFunc("/", handleJWTs)
	fmt.Println("starting server")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))

	// // debugging
	// str, err := CheckJWT("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhcHBlbmQiOiJzaGVoYXIifQ.K4hzonvWcwavW4OMYXO3IeWonCHl51RTsu07OIjUDcM", "your-256-bit-secret")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// fmt.Println("soln : ", str)
}
