package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	defaultZombieCount = 1000
	maxParallelZombies = 50
	pidFile           = "/tmp/zombie_daemon.pid"
	truePath          = "/usr/bin/true"
)

func main() {
	logFile, err := os.OpenFile("/tmp/zombie_daemon.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	if os.Geteuid() != 0 {
		log.Fatal("RUN WITH SUDO!")
	}

	if err := checkSystemLimits(); err != nil {
		log.Fatal(err)
	}

	if err := daemonize(); err != nil {
		log.Fatal(err)
	}

	if err := writePidFile(); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(pidFile)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		cancel()
	}()

	if err := createZombies(ctx, defaultZombieCount); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
	log.Println("Shutting down")
}

func daemonize() error {
	if os.Getppid() != 1 {
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Start()
		os.Exit(0)
	}
	return nil
}

func checkSystemLimits() error {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(6, &rLimit); err != nil { // 6 is RLIMIT_NPROC on Linux
		return fmt.Errorf("error getting RLIMIT_NPROC: %v", err)
	}

	if rLimit.Cur < uint64(defaultZombieCount) {
		return fmt.Errorf("RLIMIT_NPROC too low (%d). Run: ulimit -u %d", 
			rLimit.Cur, defaultZombieCount*2)
	}

	return nil
}

func writePidFile() error {
	dir := filepath.Dir(pidFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create pid file directory: %v", err)
	}

	pid := os.Getpid()
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to write pid file: %v", err)
	}

	return nil
}

func createZombies(ctx context.Context, count int) error {
	sem := make(chan struct{}, maxParallelZombies)
	defer close(sem)

	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return nil
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()

			cmd := exec.Command(truePath)
			cmd.Start()
			time.Sleep(10 * time.Millisecond)
		}()
	}
	wg.Wait()
	return nil
}