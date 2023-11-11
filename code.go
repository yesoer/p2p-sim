package main

import (
	"context"
	"fmt"
	"time"
)

// wait for ctx.Done to exit gracefully
// use fSend and fAwait to communicate between nodes
func Run(ctx context.Context, fSend func(targetId int, data any) int, fAwait func(int) int) string {
	fmt.Println("custom data ", ctx.Value("custom"))
	fmt.Println("connections ", ctx.Value("connections"))
	fmt.Println("id ", ctx.Value("id"))
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				res := fAwait(1)
				fmt.Println("res ", res)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return "done"
		default:
			time.Sleep(time.Second * 5)
			fSend(1, "data")
		}
	}
}
