package main

import (
	"context"
	"flag"
	"log"
	"module12/fibo/internal/fibo"
	"os"
	"os/signal"
	"syscall"

	cfgmaker "github.com/yukiouma/cfg-maker"
)

func main() {
	ctx := context.Background()
	cfgdir := flag.String("config", "./config.yaml", "dir of configuration")
	flag.Parse()
	cfg := cfgmaker.New(&fibo.Config{}).
		ReadFromYamlFile(*cfgdir).
		Get().(*fibo.Config)
	server := fibo.NewServer(cfg)
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
