package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type JWTPayload struct {
	Append string `json:"append"`
	Exp    int    `json:"exp,omitempty"`
	Nbf    int    `json:"nbf,omitempty"`
}

func CheckJWT(payload string, secret string) (string, error) {
	sep := strings.Split(payload, ".")
	if len(sep) != 3 {
		return "", errors.New("invalid payload")
	}

	// header
	hb, err := base64.RawURLEncoding.DecodeString(sep[0])
	if err != nil {
		return "", errors.New("error in base64-decode head: " + err.Error())
	}

	head := JWTHeader{}
	if err = json.Unmarshal(hb, &head); err != nil {
		return "", errors.New("error in JSON unmarshal head: " + err.Error())
	}

	pd, err := base64.RawURLEncoding.DecodeString(sep[1])
	if err != nil {
		return "", errors.New("error in base64-decode payload: " + err.Error())
	}

	soln := JWTPayload{}
	if err = json.Unmarshal(pd, &soln); err != nil {
		return "", errors.New("error in JSON unmarshal payload: " + err.Error())
	}

	hm := hmac.New(sha256.New, []byte(secret))
	_, err = hm.Write([]byte(sep[0] + "." + sep[1]))
	if err != nil {
		return "", errors.New("error in hmac: " + err.Error())
	}

	secretb, err := base64.RawURLEncoding.DecodeString(sep[2])
	if err != nil {
		return "", errors.New("error in base64-decode secret: " + err.Error())
	}

	if !bytes.Equal(hm.Sum(nil), secretb) {
		return "", errors.New("payload cannt be verified")
	}

	if int64(soln.Exp) > 0 && time.Now().Unix() > int64(soln.Exp) {
		return "", errors.New("expired token")
	}

	if int64(soln.Nbf) > 0 && time.Now().Unix() < int64(soln.Nbf) {
		return "", errors.New("net before yet token")
	}

	return soln.Append, nil
}
