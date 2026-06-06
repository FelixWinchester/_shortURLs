package utils

import (
	"image/png"
	"bytes"

	qrcode "github.com/skip2/go-qrcode"
)

func GenerateQRPNG(url string) ([]byte, error) {
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, qr.Image(256)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
