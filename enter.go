package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	enter "github.com/m-pavel/go-enter/lib"
)

func main() {
	interval := flag.Int("interval", 3, "Interval in seconds")
	run := flag.String("run", "", "Action to run")
	flag.Parse()
	if *run == "" {
		panic("-run is mandatory")
	}
	log.SetFlags(log.Lshortfile | log.Ldate)
	if e, err := enter.NewEnter(time.Duration(*interval)*time.Second, *run); err != nil {
		panic(err)
	} else {
		if err = e.Loop(true); err != nil {
			log.Println(err)
		}
		WaitForCtrlC(e)
	}
}

func WaitForCtrlC(e *enter.Enter) {
	var end_waiter sync.WaitGroup
	end_waiter.Add(1)
	signal_channel := make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	go func() {
		<-signal_channel
		end_waiter.Done()
	}()
	end_waiter.Wait()
	e.Close()
}
