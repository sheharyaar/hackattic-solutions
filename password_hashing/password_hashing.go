package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

type ProbScrypt struct {
	N       int    `json:"N"`
	P       int    `json:"p"`
	R       int    `json:"r"`
	Buflen  int    `json:"buflen"`
	Control string `json:"_control"`
}

type ProbPbkdf2 struct {
	Hash   string `json:"hash"`
	Rounds int    `json:"rounds"`
}

type Problem struct {
	Password string     `json:"password"`
	Salt     string     `json:"salt"`
	Pbkdf2   ProbPbkdf2 `json:"pbkdf2"`
	Scrypt   ProbScrypt `json:"scrypt"`
}

type Solution struct {
	Sha256 string `json:"sha256"`
	Hmac   string `json:"hmac"`
	Pbkdf2 string `json:"pbkdf2"`
	Scrypt string `json:"scrypt"`
}

func main() {
	probr, err := http.Get("https://hackattic.com/challenges/password_hashing/problem?access_token=85918224617cc08a")
	if err != nil {
		fmt.Println("error in getting problem : ", err)
		return
	}

	prob := Problem{}
	if err := json.NewDecoder(probr.Body).Decode(&prob); err != nil {
		fmt.Println("error in decoding json : ", err)
		return
	}

	// get salt binary
	// Note : Carefule not to use Decode(), using Decode() with a buffer pads zeroed bytes (T_T)
	// wasted 1 hour debugging this
	saltb, err := base64.StdEncoding.DecodeString(prob.Salt)
	if err != nil {
		fmt.Println("error in base64 decode : ", err)
		return
	}

	soln := Solution{}

	// sha256
	s256 := sha256.New()
	if _, err := s256.Write([]byte(prob.Password)); err != nil {
		fmt.Println("error in writitng to s256 hash : ", err)
		return
	}
	soln.Sha256 = hex.EncodeToString(s256.Sum(nil))

	// hmac
	hm := hmac.New(sha256.New, saltb)
	if _, err := hm.Write([]byte(prob.Password)); err != nil {
		fmt.Println("error in writing to hmac : ", err)
		return
	}
	soln.Hmac = hex.EncodeToString(hm.Sum(nil))

	// pbkdf2
	soln.Pbkdf2 = hex.EncodeToString(pbkdf2.Key([]byte(prob.Password), saltb,
		prob.Pbkdf2.Rounds, 32, sha256.New))

	//scrypt
	sb, err := scrypt.Key([]byte(prob.Password), saltb, prob.Scrypt.N, prob.Scrypt.R, prob.Scrypt.P, 32)
	if err != nil {
		fmt.Println("error in calculating scrypt bytes : ", err)
		return
	}
	soln.Scrypt = hex.EncodeToString(sb)

	// send solution
	solb, err := json.Marshal(&soln)
	if err != nil {
		fmt.Println("error in json marshal : ", err)
		return
	}

	buf := bytes.NewReader(solb)
	postr, _ := http.Post("https://hackattic.com/challenges/password_hashing/solve?access_token=85918224617cc08a", "application/json", buf)

	postb, _ := io.ReadAll(postr.Body)
	fmt.Println(string(postb))

}
