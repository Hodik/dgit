package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
)

func setRef(name string, hash string) {
	os.WriteFile(".dgit/" + name, []byte(hash), 0644)
}

func getRef(name string) string {
	data, err := os.ReadFile(".dgit/" + name)
	if err != nil {
    return ""
	}
	return string(data)
}

func initDgit() {
	os.MkdirAll(".dgit/objects", 0755)
	os.MkdirAll(".dgit/refs/tags", 0755)
}

func hashObject(data []byte, t string) string {
	if t == "" {
		t = "blob"
	}

	data = append(append([]byte(t), '\x00'), data...)
	h := sha1.New()
	h.Write(data)
	hash := fmt.Sprintf("%x", h.Sum(nil))
	if err := os.WriteFile(".dgit/objects/"+hash, data, 0644); err != nil {
		panic(err)
	}

	return hash
}

func catObject(hash, expected string) string {
	data, err := os.ReadFile(".dgit/objects/" + hash)
	if err != nil {
		panic(err)
	}

	parts := bytes.Split(data, []byte{'\x00'})

	if expected != "" && string(parts[0]) != expected {
		fmt.Println("Expected", expected, "but got", string(parts[0]))
    panic("Invalid object type")
	}

	return string(parts[1])
}
