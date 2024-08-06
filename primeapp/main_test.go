package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

var primeTests = []struct {
	name     string
	testNum  int
	expected bool
	msg      string
}{
	{name: "prime:7", testNum: 7, expected: true, msg: "7 is a prime number"},
	{name: "prime:4", testNum: 4, expected: false, msg: "4 is not a prime number because it is divisible by 2"},
	{name: "prime:573", testNum: 573, expected: false, msg: "573 is not a prime number because it is divisible by 3"},
	{name: "prime:0", testNum: 0, expected: false, msg: "0 is not prime by definition"},
	{name: "prime:-10", testNum: -10, expected: false, msg: "negative numbers are not prime by definition"},
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

func Test_prompt(t *testing.T) {
	//save a copy of os.Stdout
	oldOut := os.Stdout
	//create a read and write pipe
	r, w, _ := os.Pipe()
	//set os.Stdout to our write pipe
	os.Stdout = w

	prompt()

	_ = w.Close()
	//reset os.Stdout
	os.Stdout = oldOut

	out, _ := io.ReadAll(r)

	if string(out) != "-> " {
		t.Errorf("Test_prompt: incorrect prompt expected -> but got %s", string(out))
	}

}

func Test_intro(t *testing.T) {
	//save a copy of os.Stdout
	oldOut := os.Stdout
	//create a read and write pipe
	r, w, _ := os.Pipe()
	//set os.Stdout to our write pipe
	os.Stdout = w

	intro()

	_ = w.Close()
	//reset os.Stdout
	os.Stdout = oldOut

	out, _ := io.ReadAll(r)

	if !strings.Contains(string(out), "Enter a whole number") {
		t.Errorf("Test_intro: instro text not correct got: \n%s", string(out))
	}
}

var checkNumber_tests = []struct {
	name     string
	input    string
	expected string
}{
	{name: "empty", input: "", expected: "Please enter a whole number"},
	{name: "zero", input: "0", expected: "0 is not prime by definition"},
	{name: "one", input: "1", expected: "1 is not prime by definition"},
	{name: "two", input: "2", expected: "2 is a prime number"},
	{name: "four", input: "4", expected: "4 is not a prime number because it is divisible by 2"},
	{name: "negative", input: "-44", expected: "negative numbers are not prime by definition"},
	{name: "typed", input: "three", expected: "Please enter a whole number"},
	{name: "quit", input: "q", expected: ""},
	{name: "QUIT", input: "Q", expected: ""},
	{name: "decimal", input: "1.1", expected: "Please enter a whole number"},
}

func Test_checkNumbers(t *testing.T) {

	for _, e := range checkNumber_tests {
		input := strings.NewReader(e.input)
		reader := bufio.NewScanner(input)

		res, _ := checkNumbers(reader)

		if !strings.EqualFold(res, e.expected) {
			t.Errorf("%s: expected %s, but got %s", e.name, e.expected, res)
		}
	}

}

func Test_readUserInput(t *testing.T) {

	//to test this function we need a channel, and an instance of a io.reader
	doneChan := make(chan bool)

	var stdin bytes.Buffer

	stdin.Write([]byte("1\nq\n"))

	go readUserInput(&stdin, doneChan)

	<-doneChan
	close(doneChan)

}
