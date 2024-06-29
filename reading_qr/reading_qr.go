package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"net/http"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

type Resp struct {
	ImgUrl string `json:"image_url"`
}

type Solution struct {
	Code string `json:"code"`
}

func main() {

	// get image
	resp, err := http.Get("https://hackattic.com/challenges/reading_qr/problem?access_token=85918224617cc08a")
	if err != nil {
		fmt.Println("error in getting response")
	}

	v := Resp{}
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		fmt.Println("error in decoding json")
	}

	// decode QR
	rimg, err := http.Get(v.ImgUrl)
	if err != nil {
		fmt.Println("error in getting iimage binary")
	}

	img, err := png.Decode(rimg.Body)
	if err != nil {
		fmt.Println("error in decoding png : ", err)
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		fmt.Println("error in getting binary bitmap : ", err)
	}

	res, err := qrcode.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		fmt.Println("error in decoding QR : ", err)
	}
	fmt.Println("Sol : ", res.GetText())

	// send code
	pc := Solution{Code: res.GetText()}
	jb, err := json.Marshal(&pc)
	if err != nil {
		fmt.Println("error in marshaling json", err)
	}
	buf := bytes.NewBuffer(jb)

	presp, err := http.Post("https://hackattic.com/challenges/reading_qr/solve?access_token=85918224617cc08a", "application/json", buf)
	if err != nil {
		fmt.Println("error in posting solution : ", err)
	}

	pb, _ := io.ReadAll(presp.Body)
	fmt.Println(string(pb))
}
