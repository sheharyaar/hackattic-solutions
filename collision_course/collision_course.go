package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

type Problem struct {
	Include string `json:"include"`
}

type Solution struct {
	Files []string `json:"files"`
}

func main() {
	probr, err := http.Get("https://hackattic.com/challenges/collision_course/problem?access_token=85918224617cc08a")
	if err != nil {
		fmt.Println("error in GET : ", err)
		return
	}

	prob := Problem{}
	if err = json.NewDecoder(probr.Body).Decode(&prob); err != nil {
		fmt.Println("error in decoding prob json : ", err)
		return
	}

	fmt.Println("prob : ", prob)

	if err = os.WriteFile("test", []byte(prob.Include), 0644); err != nil {
		fmt.Println("error in writing test file : ", err)
		return
	}
	defer os.Remove("test")

	fmt.Println("Executing md5_fastcoll")

	cmd := exec.Command("./hashclash/bin/md5_fastcoll", "-p", "test")
	if err = cmd.Run(); err != nil {
		fmt.Println("error in execing md5_fastcoll : ", err)
		return
	}

	msg1, err := os.ReadFile("msg1.bin")
	if err != nil {
		fmt.Println("error in reading msg1 : ", err)
		return
	}
	defer os.Remove("msg1.bin")

	msg2, err := os.ReadFile("msg2.bin")
	if err != nil {
		fmt.Println("error in reading msg2 : ", err)
		return
	}
	defer os.Remove("msg2.bin")

	sol := Solution{Files: []string{base64.StdEncoding.EncodeToString(msg1),
		base64.StdEncoding.EncodeToString(msg2)}}

	fmt.Println("sol : ", sol)

	solb, err := json.Marshal(&sol)
	if err != nil {
		fmt.Println("error in soln marshal : ", err)
		return
	}

	buf := bytes.NewReader(solb)
	resp, _ := http.Post("https://hackattic.com/challenges/collision_course/solve?access_token=85918224617cc08a", "application/json", buf)

	respb, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respb))
}
