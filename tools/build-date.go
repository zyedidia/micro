package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(time.Now().Local().Format("January 02, 2006"))
}
