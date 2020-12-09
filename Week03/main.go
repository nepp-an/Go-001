package main

import (
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	done := make(chan struct{})
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, os.Interrupt)
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		return serverPProf(ctx)
	})

	g.Go(func() error {
		return server(ctx)
	})
	g.Go(func() error {
		return listenSignal(ctx, sig)
	})
	if err := g.Wait(); err != nil {
		log.Println("g.Wait error")
		log.Printf("err is %+v", err)
		close(done)
	}
	<-done
	log.Println("service down")
}

func listenSignal(ctx context.Context, sig <-chan os.Signal) error {
	select {
	case interrupt := <-sig:
		log.Println(interrupt)
		err := fmt.Errorf("os signal interrupt")
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func server(ctx context.Context) error {
	s := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: nil,
	}
	go func() {
		<-ctx.Done()
		log.Println("server shutdown")
		s.Shutdown(context.Background())
	}()
	return s.ListenAndServe()
}

func serverPProf(ctx context.Context) error {
	s := http.Server{
		Addr:    "0.0.0.0:9997",
		Handler: nil,
	}
	go func() {
		<-ctx.Done()
		log.Println("serverPProf shutdown")
		s.Shutdown(context.Background())
	}()
	return s.ListenAndServe()
}
