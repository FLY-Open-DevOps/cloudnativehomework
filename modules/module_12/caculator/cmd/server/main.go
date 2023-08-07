package main

import (
	"context"
	"flag"
	"log"
	"module12/caculator/internal/caculator"
	"os"
	"os/signal"
	"syscall"

	cfgmaker "github.com/yukiouma/cfg-maker"
)

func main() {
	ctx := context.Background()
	cfgdir := flag.String("config", "./config.yaml", "dir of configuration")
	flag.Parse()
	cfg := cfgmaker.New(&caculator.Config{}).
		ReadFromYamlFile(*cfgdir).
		Get().(*caculator.Config)
	cacu := caculator.NewCaculator(caculator.NewFibo(cfg.FiboAddr))
	server := caculator.NewServer(cfg, cacu)
	go func() {
		if err := server.Run(); err != nil {
			log.Fatalf("server down because: %v", err)
		}
	}()
	// listening SIGTERM and SIGINT
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
	if err := server.Stop(ctx); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	os.Exit(0)
}
