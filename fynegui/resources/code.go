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
	fmt.Println("out-neighbors ", ctx.Value("out-neighbors"))
	fmt.Println("in-neighbors ", ctx.Value("in-neighbors"))
	fmt.Println("id ", ctx.Value("id"))

	res := struct{ foo string }{foo: "bar"}
	inNeighbors, ok := ctx.Value("in-neighbors").([]int)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if len(inNeighbors) > 0 {
					awaitRes := fAwait(len(inNeighbors))
					fmt.Println("awaitRes ", awaitRes)
				}
			}
		}
	}()

	outNeighbors, ok := ctx.Value("out-neighbors").([]int)
	for {
		select {
		case <-ctx.Done():
			return res
		default:
			time.Sleep(time.Second * 1)
			if ok {
				for _, c := range outNeighbors {
					fmt.Println("send")
					fSend(c, "data")
				}
			}
		}
	}
}
