//go:build windows

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/sys/windows/svc"
)

const serviceName = "FingerprintQubuService"

func runWindowsService(run func(ctx context.Context) error) (bool, error) {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return false, fmt.Errorf("detecting Windows service context failed: %w", err)
	}
	if !isService {
		log.Printf("running outside Windows SCM context (foreground or wrapped by NSSM)")
		return false, nil
	}

	log.Printf("running as native Windows service: %s", serviceName)
	if err := svc.Run(serviceName, &fingerprintService{run: run}); err != nil {
		return true, fmt.Errorf("running Windows service failed: %w", err)
	}
	return true, nil
}

type fingerprintService struct {
	run func(ctx context.Context) error
}

func (s *fingerprintService) Execute(args []string, requests <-chan svc.ChangeRequest, status chan<- svc.Status) (svcSpecificExitCode bool, exitCode uint32) {
	const accepts = svc.AcceptStop | svc.AcceptShutdown

	status <- svc.Status{
		State:      svc.StartPending,
		WaitHint:   uint32((10 * time.Second).Milliseconds()),
		CheckPoint: 1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- s.run(ctx)
	}()

	status <- svc.Status{
		State:   svc.Running,
		Accepts: accepts,
	}

	for {
		select {
		case req, ok := <-requests:
			if !ok {
				continue
			}
			switch req.Cmd {
			case svc.Interrogate:
				status <- req.CurrentStatus
			case svc.Stop, svc.Shutdown:
				status <- svc.Status{
					State:   svc.StopPending,
					Accepts: accepts,
				}
				cancel()

				if err := <-serverErr; err != nil {
					log.Printf("service stop encountered error: %v", err)
					exitCode = 1
				}

				status <- svc.Status{State: svc.Stopped}
				return
			default:
				log.Printf("unsupported control request: %#v", req)
			}

		case err := <-serverErr:
			if err != nil {
				log.Printf("service runtime error: %v", err)
				exitCode = 1
			}
			status <- svc.Status{State: svc.Stopped}
			return
		}
	}
}
