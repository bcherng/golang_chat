package client

import "fmt"

func main() {
	c := make(chan string, 1) // note: len 1
	done := make(chan bool)
	counter := 0

	go inc(&counter, c, done)
	go inc(&counter, c, done)

	c <- "" // so that there's something to read from the channel (important!)

	_, _ = <-done, <-done
	fmt.Println(counter)
}
