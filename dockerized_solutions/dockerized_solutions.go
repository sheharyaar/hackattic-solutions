package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type DRegTagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type ProbCreds struct {
	Username string `json:"user"`
	Pass     string `json:"password"`
}

type Problem struct {
	Creds        ProbCreds `json:"credentials"`
	IgnitionKey  string    `json:"ignition_key"`
	TriggerToken string    `json:"trigger_token"`
}

type Trigger struct {
	RegistryHost string `json:"registry_host"`
}

type Solution struct {
	Secret string `json:"secret"`
}

func main() {
	req, err := http.Get("https://hackattic.com/challenges/dockerized_solutions/problem?access_token=85918224617cc08a")
	if err != nil {
		fmt.Println("error in GET: ", err)
		return
	}

	prob := Problem{}
	if err = json.NewDecoder(req.Body).Decode(&prob); err != nil {
		fmt.Println("error in json decode req : ", err)
		return
	}
	fmt.Printf("%v\n", prob)

	// set auth
	authCmd := exec.Command("htpasswd", "-Bb", "./registry/auth/registry.passwd", prob.Creds.Username, prob.Creds.Pass)
	if err = authCmd.Run(); err != nil {
		fmt.Println("error in running htpasswd: ", err)
		return
	}

	fmt.Println("waiting for 2 seconds for you to run docker compose")
	time.Sleep(2 * time.Second)

	fmt.Println("sendiing endpoint to hackattic")
	registry := Trigger{
		RegistryHost: "6b8a-2405-201-a42a-e05e-f049-d2bc-838f-d7b9.ngrok-free.app",
	}

	trigb, err := json.Marshal(&registry)
	if err != nil {
		fmt.Println("error in json Marshal trigger :", err)
		return
	}

	buf := bytes.NewReader(trigb)
	_, err = http.Post(fmt.Sprintf("https://hackattic.com/_/push/%s", prob.TriggerToken), "application/json", buf)
	if err != nil {
		fmt.Println("error in sending trigger: ", err)
		return
	}

	// trigrespb, _ := io.ReadAll(trigresp.Body)
	// fmt.Printf("%s\n", string(trigrespb))

	cl := http.Client{}
	dreq, _ := http.NewRequest("GET", "https://6b8a-2405-201-a42a-e05e-f049-d2bc-838f-d7b9.ngrok-free.app/v2/hack/tags/list", nil)
	dreq.SetBasicAuth(prob.Creds.Username, prob.Creds.Pass)

	tag, err := cl.Do(dreq)
	if err != nil {
		fmt.Println("error in GET tags:", err)
		return
	}

	tags := DRegTagList{}
	json.NewDecoder(tag.Body).Decode(&tags)
	fmt.Printf("tags : %v\n", tags.Tags)

	// docker login
	loginCmd := exec.Command("docker", "login", "6b8a-2405-201-a42a-e05e-f049-d2bc-838f-d7b9.ngrok-free.app", "-u", prob.Creds.Username, "-p", prob.Creds.Pass)
	if err := loginCmd.Run(); err != nil {
		fmt.Println("error in docker login : ", err)
		return
	}

	for _, tag := range tags.Tags {
		pullStr := fmt.Sprintf("6b8a-2405-201-a42a-e05e-f049-d2bc-838f-d7b9.ngrok-free.app/hack:%s", tag)
		fmt.Println("pull cmd :", pullStr)

		dockCmd := exec.Command("docker", "pull", pullStr)
		if err := dockCmd.Run(); err != nil {
			fmt.Println("error in runniing docker pull :", err)
			return
		}

		// just to let everything get settled
		time.Sleep(1 * time.Millisecond)

		// docker run it with env
		imgStr := "6b8a-2405-201-a42a-e05e-f049-d2bc-838f-d7b9.ngrok-free.app/hack:" + tag
		ignStr := fmt.Sprintf("IGNITION_KEY=%s", prob.IgnitionKey)

		fmt.Println("img cmd :", imgStr)
		fmt.Println("ignition cmd :", ignStr)

		dockexecCmd := exec.Command("docker", "run", "-e", ignStr, imgStr)
		outb, err := dockexecCmd.Output()
		if err != nil {
			fmt.Println("error in running docker run : ", err)
			return
		}

		fmt.Printf("tried :%s, out: %s\n", tag, string(outb))

		if !strings.Contains(string(outb), "oops") {
			soln := Solution{Secret: strings.TrimSuffix(string(outb), "\n")}
			solb, err := json.Marshal(&soln)
			if err != nil {
				fmt.Println("error in JSON marshal: ", err)
				return
			}

			buf := bytes.NewReader(solb)

			post, err := http.Post("https://hackattic.com/challenges/dockerized_solutions/solve?access_token=85918224617cc08a", "application/json", buf)
			if err != nil {
				fmt.Println("error in POST soln : ", err)
				return
			}

			postb, _ := io.ReadAll(post.Body)
			fmt.Println(string(postb))
			return
		}
	}
}
