package main

import (
	"fmt"
)

func main() {
	race()
}

func race() {

	wait := make(chan struct{})

	type test struct {
		value1 int
		value2 int
	}

	t := &test{
		value1: 1,
		value2: 2,
	}
	go func() {
		fmt.Println("START1")
		for i := 0; i < 1000000; i++ {
			t.value1++ // read, increment, write
		}
		fmt.Println("DONE1")
		close(wait)
	}()
	fmt.Println("START")
	for i := 0; i < 1000000; i++ {
		t.value2++ // read, increment, write
	}
	fmt.Println("DONE")
	<-wait
	fmt.Println(t) // Output: <unspecified>
}
