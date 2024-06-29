package main

/*
#cgo LDFLAGS: -lzip
#include <zip.h>
#include <stdlib.h>
*/
import "C"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type RespZip struct {
	ZipUrl string `json:"zip_url"`
}

type SolZip struct {
	Secret string `json:"secret"`
}

func main() {
	// fetch the zip file
	rzip, err := http.Get("https://hackattic.com/challenges/brute_force_zip/problem?access_token=85918224617cc08a")
	if err != nil {
		fmt.Println("error in fetching zip json : ", err)
		return
	}
	defer rzip.Body.Close()

	rb := RespZip{}
	err = json.NewDecoder(rzip.Body).Decode(&rb)
	if err != nil {
		fmt.Println("error in decoding zip json : ", err)
		return
	}

	rzipb, err := http.Get(rb.ZipUrl)
	if err != nil {
		fmt.Println("error in fetching zip file : ", err)
		return
	}
	defer rzipb.Body.Close()

	zipb, err := io.ReadAll(rzipb.Body)
	if err != nil {
		fmt.Println("error in reading zip bytes : ", err)
	}

	err = os.WriteFile("task", zipb, 0644)
	if err != nil {
		fmt.Println("error in writing to file")
	}

	fmt.Println("executing fcrackzip")

	cmd := exec.Command("fcrackzip", "-b", "-c", "a1", "-l", "4-6", "-u", "task")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("error in stdoutpipe :", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("error in stdoutpipe :", err)
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("error in starting fcrackzip : ", err)
	}

	pyb, err := io.ReadAll(stdout)
	if err != nil {
		fmt.Println("error in reading stdout : ", err)
	}

	pybe, err := io.ReadAll(stderr)
	if err != nil {
		fmt.Println("error in reading stderr : ", err)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println("error in wait : ", err)
	}

	fmt.Println(pyb)
	fmt.Println(string(pybe))

	s := strings.Split(string(pyb), "==")
	s[1], _ = strings.CutPrefix(s[1], " ")
	s[1], _ = strings.CutSuffix(s[1], "\n")

	// unzip the archive and extract password
	pass := "-p" + s[1]
	cmd2 := exec.Command("7z", "x", pass, "task")
	err = cmd2.Run()
	if err != nil {
		fmt.Println("error in unzipping file : ", err)
	}

	//read file
	b, err := os.ReadFile("secret.txt")
	if err != nil {
		fmt.Println("error in reading secret")
	}

	b = []byte(strings.TrimSuffix(string(b), "\n"))

	sol := SolZip{Secret: string(b)}
	solb, _ := json.Marshal(&sol)
	buf := bytes.NewBuffer(solb)

	r, err := http.Post("https://hackattic.com/challenges/brute_force_zip/solve?access_token=85918224617cc08a", "application/json", buf)
	if err != nil {
		fmt.Println("error in POST")
	}

	br, _ := io.ReadAll(r.Body)
	fmt.Println(string(br))
}
