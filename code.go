package main

import (
	"context"
	"fmt"
	"time"
)

type sendFunc func(targetId int, data any) int
type awaitFunc func(int) []any

// wait for ctx.Done to exit gracefully
// use fSend and fAwait to communicate between nodes
func Run(ctx context.Context, fSend sendFunc, fAwait awaitFunc) any {
	fmt.Println("custom data ", ctx.Value("custom"))
	fmt.Println("neighbors ", ctx.Value("outgoing-connections"))
	fmt.Println("id ", ctx.Value("id"))
	res := struct{ foo string }{foo: "bar"}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				awaitRes := fAwait(1)
				fmt.Println("awaitRes ", awaitRes)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return res
		default:
			time.Sleep(time.Second * 1)
			neighbors, ok := ctx.Value("outgoing-connections").([]int)
			if ok {
				fmt.Println("send")
				for _, c := range neighbors {
					fSend(c, "data")
				}
			}
		}
	}
}
