package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
)

type Response struct {
	Byte string `json:"bytes"`
}

type Unpack struct {
	Si  int32   `json:"int"`
	Ui  uint32  `json:"uint"`
	Sh  int16   `json:"short"`
	Fl  float32 `json:"float"`
	Do  float64 `json:"double"`
	BDo float64 `json:"big_endian_double"`
}

func main() {
	resp, err := http.Get("https://hackattic.com/challenges/help_me_unpack/problem?access_token=85918224617cc08a")
	if err != nil {
		fmt.Println("error in getting string")
	}

	// bytes -> json -> base64 -> bytes -> struct ->  json
	v := Response{}
	u := Unpack{}

	// byte -> json
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		fmt.Println("error in decoding request")
	}

	// json -> base64
	b := make([]byte, 32)
	b, err = base64.StdEncoding.DecodeString(v.Byte)
	if err != nil {
		fmt.Println("error in decoding base64 -> bytes")
	}

	fmt.Println(v.Byte)

	// binary.Read(bytes.NewReader(b), binary.NativeEndian, &u)
	u.Si = int32(binary.NativeEndian.Uint32(b[0:4]))
	u.Ui = binary.NativeEndian.Uint32(b[4:8])
	u.Sh = int16(binary.NativeEndian.Uint16(b[8:10]))
	u.Fl = math.Float32frombits(binary.NativeEndian.Uint32(b[12:16]))
	u.Do = math.Float64frombits(binary.NativeEndian.Uint64(b[16:24]))
	u.BDo = math.Float64frombits(binary.BigEndian.Uint64(b[24:32]))

	fmt.Println(b)
	var c []byte
	buf := bytes.NewBuffer(c)
	err = json.NewEncoder(buf).Encode(&u)
	if err != nil {
		fmt.Println("error in encoding to json : ", err)
	}

	fmt.Println(u)

	nresp, err := http.Post("https://hackattic.com/challenges/help_me_unpack/solve?access_token=85918224617cc08a", "application/json", buf)
	if err != nil {
		fmt.Println("error in post")
	}

	x, err := io.ReadAll(nresp.Body)
	if err != nil {
		fmt.Println("errr in read post")
	}
	fmt.Println(string(x))

}
