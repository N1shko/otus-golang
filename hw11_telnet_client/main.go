package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "timeout value, e.g. 10s")
	flag.Parse()
	if flag.NArg() != 2 {
		log.Fatalf("Exactly 2 arguments required, you provided %d", flag.NArg())
	}
	if _, err := strconv.Atoi(flag.Arg(1)); err != nil {
		log.Fatalf("Port value must be int, you provided: %s", flag.Arg(1))
	}
	client := NewTelnetClient(net.JoinHostPort(flag.Arg(0), flag.Arg(1)), *timeout, os.Stdin, os.Stdout)
	if err := client.Connect(); err != nil {
		log.Fatalln(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP)
	defer cancel()
	go func() {
		if err := client.Receive(); err != nil {
			log.Print(err)
		}
		cancel()
	}()
	go func() {
		if err := client.Send(); err != nil {
			log.Print(err)
		}
		cancel()
	}()
	<-ctx.Done()
}
