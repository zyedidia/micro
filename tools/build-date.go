//go:build ignore

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	var buildTime time.Time
	epoch := os.Getenv("SOURCE_DATE_EPOCH")
	if epoch != "" {
		i, err := strconv.Atoi(epoch)
		if err != nil {
			fmt.Errorf("SOURCE_DATE_EPOCH is not a valid integer")
			os.Exit(1)
		}
		buildTime = time.Unix(int64(i), 0)
	} else {
		buildTime = time.Now().Local()
	}
	fmt.Println(buildTime.Format("January 02, 2006"))
}
