package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {

	// Print a welcome message
	intro()
	// create a channel to indicate when a user wants to quit
	doneChan := make(chan bool)
	// start a go routine to read user input and run program
	go readUserInput(doneChan)
	//block until the doneChan gets a value
	<-doneChan
	// close the channel
	close(doneChan)
	//say goodbye
	fmt.Println("Goodbye...")
}

func intro() {
	fmt.Println("Is it prime?")
	fmt.Println("------------")
	fmt.Println("Enter a whole number, and we'll tell you if it is a prime number or not. Enter q to quit.")
	prompt()
}

func prompt() {
	fmt.Print("-> ")
}

func readUserInput(doneChan chan bool) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		res, done := checkNumbers(scanner)

		if done {
			doneChan <- true
			return
		}

		fmt.Println(res)
		prompt()
	}
}

func checkNumbers(scanner *bufio.Scanner) (string, bool) {
	//read user input
	scanner.Scan()

	//check to see if user wants to quit
	if strings.EqualFold(scanner.Text(), "q") {
		return "", true
	}

	numToCheck, err := strconv.Atoi(scanner.Text())

	if err != nil {
		return "Please enter a whole number", false
	}

	_, msg := isPrime(numToCheck)
	return msg, false
}

func isPrime(n int) (bool, string) {

	if n == 0 || n == 1 {
		return false, fmt.Sprintf("%d is not prime by definition", n)
	}

	if n < 0 {
		return false, "negative numbers are not prime by definition"
	}

	for i := 2; i <= n/2; i++ {
		if n%i == 0 {
			return false, fmt.Sprintf("%d is not a prime number because it is divisible by %d", n, i)
		}
	}

	return true, fmt.Sprintf("%d is a prime number", n)

}
