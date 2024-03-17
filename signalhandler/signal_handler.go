package signalhandler

import (
	"log/slog"
	"os"
	"os/signal"
)

type SignalHandler struct {
	signals   []os.Signal
	signalsCh chan os.Signal
}

func New(signals ...os.Signal) *SignalHandler {
	return &SignalHandler{
		signals:   signals,
		signalsCh: make(chan os.Signal, 1),
	}
}

func (sh *SignalHandler) Run() error {
	signal.Notify(sh.signalsCh, sh.signals...)
	s, ok := <-sh.signalsCh
	if !ok {
		return nil
	}

	slog.Info("Signal received", slog.String("signal", s.String()))
	return nil
}

func (sh *SignalHandler) Stop() error {
	close(sh.signalsCh)
	return nil
}
