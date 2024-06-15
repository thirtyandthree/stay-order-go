package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func GetOrderNo() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf(
		"%s%d",
		time.Now().Format("20060102150405"),
		r.Intn(100000),
	)
}
