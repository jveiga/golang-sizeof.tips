package app

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"
	// "github.com/gophergala/golang-sizeof.tips/internal/log"
)

func Run() (exitCode int) {
	// switch isDaemon, err := daemon.Daemonize(); {
	// case !isDaemon:
	// 	return
	// case err != nil:
	// 	log.StdErr("could not start daemon, reason -> %s", err.Error())
	// 	return 1
	// }

	var err error
	// appLog, err = log.NewApplicationLogger()
	if err != nil {
		log.Println("could not create access log, reason -> %s", err.Error())
		return 1
	}

	if err = prepareTemplates(); err != nil {
		log.Println("could not parse html templates, reason -> %s", err.Error())
		return 1
	}

	httpPort := os.Getenv("_GO_HTTP")
	if httpPort == "" {
		httpPort = DefaultHttpPort
	}

	bindHttpHandlers()
	canExit, httpErr := make(chan sig, 1), make(chan error, 1)
	go func() {
		defer close(canExit)
		if err := http.ListenAndServe(httpPort, nil); err != nil {
			httpErr <- errors.New(fmt.Sprintf("creating HTTP server on port '%s' FAILED, reason -> %s\n", httpPort, err.Error()))
		}
	}()
	select {
	case err = <-httpErr:
		// appLog.Error(err.Error())
		log.Println(err.Error())
		return 1
	case <-time.After(300 * time.Millisecond):
	}

	notifyParentProcess()

	<-canExit
	return
}

// Notifies parent process that everything is OK.
func notifyParentProcess() {
	if err := syscall.Kill(os.Getppid(), syscall.SIGUSR1); err != nil {
		log.Println(
			"Notifying parent process FAILED, reason -> %s", err.Error(),
		)
	} else {
		log.Println("Notifying parent process SUCCEED")
	}
}
