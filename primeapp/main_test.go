package main

import "testing"

var primeTests = []struct {
	name     string
	testNum  int
	expected bool
	msg      string
}{
	{name: "prime-7", testNum: 7, expected: true, msg: "7 is a prime number"},
	{name: "prime-4", testNum: 4, expected: false, msg: "4 is not a prime number because it is divisible by 2"},
	{name: "prime-573", testNum: 573, expected: false, msg: "573 is not a prime number because it is divisible by 3"},
}

func Test_isPrime(t *testing.T) {

	for _, e := range primeTests {

		result, msg := isPrime(e.testNum)

		if result && !e.expected {
			t.Errorf("%s: with %d as test param, got true, but expected false", e.name, 0)
		}

		if !result && e.expected {
			t.Errorf("%s: with %d as test param, got false, but expected true", e.name, 0)
		}

		if msg != e.msg {
			t.Errorf("%s: wrong message returned expected: %s but got: %s", e.name, e.msg, msg)
		}
	}

}
