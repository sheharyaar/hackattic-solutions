package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type BlockData struct {
	Str string
	Num int
}

type Block struct {
	Data  []interface{} `json:"data"`
	Nonce int           `json:"nonce"`
}

type Problem struct {
	Difficulty int   `json:"difficulty"`
	Blk        Block `json:"block"`
}

type Solution struct {
	Nonce int `json:"nonce"`
}

func CheckSol(hash [32]byte, diff int) bool {
	var i int
	for i = 0; i < diff/8; i++ {
		if hash[i] != 0 {
			return false
		}
	}

	val := hash[i] >> (8 - diff%8) & 0xff
	if val != 0 {
		return false
	}

	return true
}

func main() {
	probr, err := http.Get("https://hackattic.com/challenges/mini_miner/problem?access_token=85918224617cc08a")
	if err != nil {
		fmt.Println("error in getting problem resp : ", err)
		return
	}

	prob := Problem{}
	if err = json.NewDecoder(probr.Body).Decode(&prob); err != nil {
		fmt.Println("error in decoding prob json : ", err)
		return
	}

	fmt.Println("starting to calculate")

	var nonce int = 0
	var hash [32]byte

	testjson := Block{Data: prob.Blk.Data, Nonce: nonce}
	for {

		testjson.Nonce = nonce
		tb, err := json.Marshal(&testjson)
		if err != nil {
			fmt.Println("error in marshaling test json : ", err)
		}

		hash = sha256.Sum256(tb)
		if CheckSol(hash, prob.Difficulty) {
			break
		}
		nonce++
	}

	fmt.Println("sol : ", nonce)

	sol := Solution{Nonce: nonce}
	solb, err := json.Marshal(&sol)
	if err != nil {
		fmt.Println("error in marshalling json : ", err)
	}

	buf := bytes.NewReader(solb)

	postr, err := http.Post("https://hackattic.com/challenges/mini_miner/solve?access_token=85918224617cc08a", "application/json", buf)
	if err != nil {
		fmt.Println("error in POST : ", err)
	}

	postb, _ := io.ReadAll(postr.Body)
	fmt.Println(string(postb))

}
