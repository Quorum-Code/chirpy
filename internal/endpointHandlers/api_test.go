package endpointhandlers

import (
	"fmt"
	"testing"
)

func TestFunc(t *testing.T) {
	fmt.Println("ran a test")

	a := 1 + 2

	if a != 3 {
		t.Error("oh no")
	}
}
