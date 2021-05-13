package main

import (
	"fmt"
	"sync"
	//	"strconv"
)

func printNum(i int, wg *sync.WaitGroup) {
	fmt.Println("i: i", i)
	wg.Done()
}

func main() {
	/*
		A := "1111"
		B := "1000"
		C := "0011"
		D := "0101"

		fmt.Printf("A | B: %b\n", strconv.ParseInt(A,2,0) | strconv.ParseInt(A,2,0))
		fmt.Printf("C & D: %b\n", strconv.ParseInt(C) & strconv.ParseInt(D))
		fmt.Println("Hello, playground")
		fmt.Printf("%b",5 | 1)*/

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		go printNum(i, &wg)
		wg.Add(1)
	}
	wg.Wait()
}
