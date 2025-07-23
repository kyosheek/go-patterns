package main

import (
	"fmt"
	"time"

	"github.com/kyosheek/go-patterns/pkg/throttle"
)

func main() {
	fn := func(args ...any) {
		fmt.Println(time.Now().Format("15:04:05.000"), args)
	}

	throttled := throttle.New(fn, 250*time.Millisecond)

	for i := 0; i < 10; i++ {
		throttled(i)
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(1 * time.Second)
}
