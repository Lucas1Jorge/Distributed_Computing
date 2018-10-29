package main

import (
	// "fmt"
	"golang.org/x/tour/wc"
	"strings"
)

func WordCount(s string) map[string]int {
	ans := make(map[string]int)
	
	for _, word := range strings.Split(s, " ") {
		count, ok := ans[word]
		if ok {
			ans[word] = count + 1
		} else {
			ans[word] = 1
		}
	}
	
	return ans
}

func main() {
	wc.Test(WordCount)
}
