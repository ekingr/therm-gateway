package main

import (
	"crypto/rand"
	b64 "encoding/base64"
	"fmt"
)

// Generate auth key for wwwHome/backAuth/authConfig.json

const keyLen = 32

func main() {
	k := make([]byte, keyLen)
	_, err := rand.Read(k)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(b64.StdEncoding.EncodeToString(k))
}
