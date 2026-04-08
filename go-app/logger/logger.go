package logger

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
)

const LOGGER_BUFFER_SIZE = 128

type GlobalLogger_Event struct {
	LogLevel string
	Message  string
}

type GlobalLogger struct {
	mutex     sync.RWMutex
	slogger   *slog.Logger
	listeners []chan GlobalLogger_Event
}

func (g *GlobalLogger) Listen() (chan GlobalLogger_Event, func()) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	channel := make(chan GlobalLogger_Event, LOGGER_BUFFER_SIZE)
	g.listeners = append(g.listeners, channel)
	unsubscribe := func() {
		g.mutex.Lock()
		defer g.mutex.Unlock()

		for index, c := range g.listeners {
			if c == channel {
				g.listeners = append(g.listeners[:index], g.listeners[index+1:]...)
				break
			}
		}
		close(channel)
	}
	return channel, unsubscribe
}

func (g *GlobalLogger) PropertiesFromArgs(args ...any) map[string]string {
	properties := map[string]string{}
	for index, arg := range args {
		if index%2 == 1 {
			/* uneven indexes are the values */
			key := fmt.Sprintf("%v", args[index-1])
			if as_err, ok := arg.(error); ok {
				properties[key] = fmt.Sprintf("%#v", as_err.Error())
			} else {
				properties[key] = fmt.Sprintf("%#v", arg)
			}
		}
	}
	return properties
}

func (g *GlobalLogger) Debug(msg string, args ...any) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.slogger.Debug(msg, args...)

	if len(g.listeners) > 0 {
		properties := g.PropertiesFromArgs(args...)
		for _, c := range g.listeners {
			c <- GlobalLogger_Event{LogLevel: "debug", Message: fmt.Sprintf("%s | %v", msg, properties)}
		}
	}
}

func (g *GlobalLogger) Info(msg string, args ...any) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.slogger.Info(msg, args...)

	if len(g.listeners) > 0 {
		properties := g.PropertiesFromArgs(args...)
		for _, c := range g.listeners {
			c <- GlobalLogger_Event{LogLevel: "info", Message: fmt.Sprintf("%s | %v", msg, properties)}
		}
	}
}

func (g *GlobalLogger) Error(msg string, args ...any) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.slogger.Error(msg, args...)

	if len(g.listeners) > 0 {
		properties := g.PropertiesFromArgs(args...)
		for _, c := range g.listeners {
			c <- GlobalLogger_Event{LogLevel: "error", Message: fmt.Sprintf("%s | %v", msg, properties)}
		}
	}
}

var Logger = GlobalLogger{
	mutex: sync.RWMutex{},
	slogger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})),
	listeners: []chan GlobalLogger_Event{},
}
