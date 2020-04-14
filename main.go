package main

import (
	"os"
	"os/signal"

	"etcord/server"

	log "github.com/sirupsen/logrus"
)

func main() {
	// TODO config
	if len(os.Args) != 2 {
		log.Error("Missing port number")
		return
	}

	log.SetLevel(log.TraceLevel)

	// TODO
	s := server.NewServer(os.Args[1])
	s.AddChannel()
	s.Start()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		select {
		case <-sig:
			s.Stop()
		}

		for {
			select {
			case <-sig:
				log.Info("Already shutting down")
			}
		}
	}()

	s.Wait()
}
