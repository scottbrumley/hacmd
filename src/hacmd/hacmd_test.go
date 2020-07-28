package hacmd

import (
	"fmt"
	"strings"
	"testing"
)

// Test CheckAPI Function
func TestCheckAPI(t *testing.T) {

	fmt.Println("Test #1: CheckAPI empty host")
	// Test for empty argument
	emptyResult := CheckAPI("")

	if emptyResult != " Not a valid IP Address or Host Name" {
		t.Errorf("checkAPI(\"\") failed, expected %v, got %v", " Not a valid IP Address or Host Name", emptyResult)
	}

	fmt.Println("Test #2: CheckAPI Bad IP")
	// Test bad IP Address
	badIP := CheckAPI("192.168.192.999")

	if badIP != "192.168.192.999 Not a valid IP Address or Host Name" {
		t.Errorf("checkAPI(\"\") failed, expected %v, got %v", "192.168.192.199 Not a valid IP Address or Host Name", emptyResult)
	}

	fmt.Println("Test #3: CheckAPI Invalid Host")
	// Test bad IP Address
	badHost := CheckAPI("marvel@lab.com")

	if badHost != "marvel@lab.com Not a valid IP Address or Host Name" {
		t.Errorf("checkAPI(\"\") failed, expected %v, got %v", "marvel@lab.com Not a valid IP Address or Host Name", emptyResult)
	}

	fmt.Println("Test #4: CheckAPI connection refused")
	// Test bad IP Address
	badHost = CheckAPI("127.0.0.1")

	if !strings.Contains(badHost, "dial tcp 127.0.0.1:443: connect: connection refused") {
		t.Errorf("checkAPI(\"\") failed, expected %v, got %v", "dial tcp 127.0.0.1:443: connect: connection refused", emptyResult)
	}

}
