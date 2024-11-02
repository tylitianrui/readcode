package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oklog/run"
)

func main() {
	//  run.Group
	var g run.Group

	term := make(chan os.Signal, 1)

	cancel := make(chan struct{})
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	time1 := NewXtimer("g2")
	time2 := NewXtimer("g3")

	//
	g.Add(
		func() error {
			select {
			case sig := <-term:
				fmt.Println("g1接收到系统信号", sig.String())
				return fmt.Errorf("g1接收到系统信号 退出")
			case <-cancel:
				fmt.Println("cancel 有信号了")
			}
			return nil
		},
		func(err error) {
			fmt.Println("g1 --interrupt函数执行")
			close(cancel)
		},
	)

	g.Add(
		time1.PrintTime, time1.Stop,
	)
	g.Add(
		time2.PrintTime, time2.Stop,
	)

	if err := g.Run(); err != nil {
		fmt.Println("程序退出。。。")
		os.Exit(1)
	}

}

type Xtimer struct {
	Name   string
	ctx    context.Context
	cancel context.CancelFunc
}

func NewXtimer(name string) *Xtimer {
	ctx, cancel := context.WithCancel(context.TODO())
	return &Xtimer{
		Name:   name,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (t *Xtimer) PrintTime() error {
	for {
		select {
		case <-t.ctx.Done():
			fmt.Println(t.Name, "退出")
			return fmt.Errorf("%v stop", t.Name)
		default:
			time.Sleep(2 * time.Second)
			fmt.Println(t.Name, time.Now())
		}
	}

}

func (t *Xtimer) Stop(err error) {
	t.cancel()
	fmt.Println(t.Name, "--interrupt函数执行")
}
