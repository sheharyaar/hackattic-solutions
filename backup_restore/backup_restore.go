package main

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	_ "github.com/lib/pq"
)

type Problem struct {
	Dump string `json:"dump"`
}

type Soln struct {
	AliveSSNs []string `json:"alive_ssns"`
}

func main() {
	resp, err := http.Get("https://hackattic.com/challenges/backup_restore/problem?access_token=85918224617cc08a")
	if err != nil {
		fmt.Println("error in getting the dump")
	}

	problem := Problem{}
	err = json.NewDecoder(resp.Body).Decode(&problem)
	if err != nil {
		fmt.Println("error iin decoding challenge json : ", err)
	}

	dumpb, err := base64.StdEncoding.DecodeString(problem.Dump)
	if err != nil {
		fmt.Println("error in decoding base64 : ", err)
	}

	gread, err := gzip.NewReader(bytes.NewReader(dumpb))
	if err != nil {
		fmt.Println("error in creating gzip reader : ", err)
	}

	gzb, err := io.ReadAll(gread)
	if err != nil {
		fmt.Println("error in reading gzip bytes : ", err)
	}

	err = os.WriteFile("pgdump.sql", gzb, 0644)
	if err != nil {
		fmt.Println("error in writing dump bytes to file : ", err)
	}

	// load dump into postgres
	cmd := exec.Command("psql", "-d", "dump", "-f", "pgdump.sql")
	err = cmd.Run()
	if err != nil {
		fmt.Println("error in execing psql : ", err)
	}

	// run query
	connStr := "user=wazir dbname=dump sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	status := "alive"
	rows, err := db.Query("SELECT ssn FROM criminal_records WHERE status = $1", status)
	if err != nil {
		fmt.Println("error in executing query : ", err)
		return
	}
	defer rows.Close()

	ssns := make([]string, 0)
	for rows.Next() {
		var ssn string
		if err := rows.Scan(&ssn); err != nil {
			fmt.Println("error in scanning row : ", err)
			return
		}

		ssns = append(ssns, ssn)
	}

	soln := Soln{AliveSSNs: ssns}
	solb, err := json.Marshal(&soln)
	if err != nil {
		fmt.Println("error in marshalling json : ", err)
	}

	buf := bytes.NewReader(solb)
	solresp, err := http.Post("https://hackattic.com/challenges/backup_restore/solve?access_token=85918224617cc08a", "application/json", buf)
	if err != nil {
		fmt.Println("error in posting : ", err)
	}

	solrb, _ := io.ReadAll(solresp.Body)
	fmt.Println(string(solrb))
}
