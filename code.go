package main

import (
	"fmt"
	"time"
)

func Run(fSend func(targetId int, data any) int, fAwait func(int) int) string {
	go func() {
		for {
			res := fAwait(1)
			fmt.Println("res ", res)
		}
	}()

	for {

		time.Sleep(time.Second * 5)
		fSend(1, "data")
	}
	return "done"
}
