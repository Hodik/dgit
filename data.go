package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"strings"
)

type RefValue struct {
	value      string
	isSymbolic bool
}

func setRef(name string, value *RefValue, deref bool) {
	if value.value == "" {
		panic("Invalid value")
	}

	ref, _ := getRefInternal(name, deref)
	if value.isSymbolic {
		os.WriteFile(".dgit/"+ref, []byte("ref: "+value.value), 0644)
	} else {
		os.WriteFile(".dgit/"+ref, []byte(value.value), 0644)
	}
}

func getRef(name string, deref bool) *RefValue {
	_, ref := getRefInternal(name, deref)
	return ref
}

func getRefInternal(name string, deref bool) (string, *RefValue) {
	data, err := os.ReadFile(".dgit/" + name)
	if err != nil {
		return name, nil
	}

	strdata := string(data)
	symbolic := strings.HasPrefix(strdata, "ref: ")
	if symbolic {
		strdata = strings.TrimSpace(strings.Split(strdata, ":")[1])
		if deref {
			return getRefInternal(strings.TrimPrefix(strdata, "ref: "), deref)
		}
	}

	return name, &RefValue{value: strdata, isSymbolic: symbolic}
}

func initDgit() {
	os.MkdirAll(".dgit/objects", 0755)
	os.MkdirAll(".dgit/refs/tags", 0755)
	os.MkdirAll(".dgit/refs/heads", 0755)
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
