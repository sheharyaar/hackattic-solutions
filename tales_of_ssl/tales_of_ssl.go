package main

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/biter777/countries"
)

type ProbReqData struct {
	Domain       string `json:"domain"`
	SerialNumber string `json:"serial_number"`
	Country      string `json:"country"`
}

type Problem struct {
	PrivateKey string      `json:"private_key"`
	ReqData    ProbReqData `json:"required_data"`
}

type Solution struct {
	Certificate string `json:"certificate"`
}

func main() {
	req, err := http.Get("https://hackattic.com/challenges/tales_of_ssl/problem?access_token=85918224617cc08a")
	if err != nil {
		fmt.Println("error in GET: ", err)
		return
	}

	prob := Problem{}
	if err = json.NewDecoder(req.Body).Decode(&prob); err != nil {
		fmt.Println("error in decoding req json :", err)
		return
	}

	fmt.Println("prob : ", prob)

	baseb, err := base64.StdEncoding.DecodeString(prob.PrivateKey)
	if err != nil {
		fmt.Println("error in base64 decode :", err)
		return
	}

	privkey, err := x509.ParsePKCS1PrivateKey(baseb)
	if err != nil {
		fmt.Println("error in parsing private key : ", err)
		return
	}

	cntry := countries.ByName(prob.ReqData.Country)
	subj := pkix.Name{
		Country:      []string{cntry.Alpha2()},
		SerialNumber: prob.ReqData.SerialNumber,
		CommonName:   prob.ReqData.Domain,
	}

	snob, err := hex.DecodeString(strings.TrimPrefix(prob.ReqData.SerialNumber, "0x"))
	if err != nil {
		fmt.Println("error in decoding hex : ", err)
	}

	sno := int64(binary.BigEndian.Uint32(snob))
	fmt.Println("raw: ", binary.BigEndian.Uint32(snob))
	fmt.Println("searial : ", sno)
	xCert := x509.Certificate{
		SerialNumber:       big.NewInt(sno),
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
		NotBefore:          time.Now(),
		NotAfter:           time.Now().AddDate(10, 0, 0),
	}

	der, err := x509.CreateCertificate(rand.Reader, &xCert, &xCert, privkey.Public(), privkey)
	if err != nil {
		fmt.Println("error in generating certificate : ", err)
		return
	}

	fmt.Println(xCert)
	cert := base64.StdEncoding.EncodeToString(der)

	soln := Solution{Certificate: cert}
	solb, err := json.Marshal(&soln)
	if err != nil {
		fmt.Println("error in marshaling json: ", err)
		return
	}
	buf := bytes.NewReader(solb)
	postr, _ := http.Post("https://hackattic.com/challenges/tales_of_ssl/solve?access_token=85918224617cc08a", "application/json", buf)
	postb, _ := io.ReadAll(postr.Body)
	fmt.Println(string(postb))
}
