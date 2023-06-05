package module1

import (
	"log"
	"strings"
	"testing"
)

func Test1_1(t *testing.T) {
	const singleBlank = " "
	slice := []string{"I", "am", "stupid", "and", "weak"}
	expect := []string{"I", "am", "smart", "and", "strong"}
	log.Printf("origin slice is %v", slice)
	ModifySlice(slice)
	if strings.Join(slice, singleBlank) != strings.Join(expect, singleBlank) {
		t.Fatalf("expect %v, got %v", expect, slice)
	}
	log.Printf("modified slice is %v", slice)
}
