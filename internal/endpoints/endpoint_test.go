package endpoints

import (
	"log"
	"testing"
)

func TestFunc(t *testing.T) {
	log.Println("ran a test")

	a := 1 + 2

	if a != 3 {
		t.Error("oh no")
	}
}

func TestAnotherOne(t *testing.T) {
	log.Println("running another test")

	c := 0

	if c != 0 {
		t.Error("ummmm")
	}
}
