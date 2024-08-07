package services

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"oh-my-chat/src/config"
	"oh-my-chat/src/core"
	"oh-my-chat/src/logger"
)

type StorageService interface {
	Dequeue() (core.ActionReplyPair, bool)
}

type Worker struct {
	storage StorageService
}

func (w *Worker) Produce(ctx context.Context, action chan<- core.ActionReplyPair) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			actionPair, ok := w.storage.Dequeue()
			if ok {
				action <- actionPair
			}

		}
	}
}

func (w *Worker) Consume(ctx context.Context, action <-chan core.ActionReplyPair) {
	workerLog := logger.Logger.With(zap.String("context", "worker"))
	for {
		select {

		case <-ctx.Done():
			return

		case actionPair := <-action:
			err := actionPair.Action.Handle(ctx, &actionPair.Input)
			if err != nil {
				workerLog.Error("Error Handling Action", zap.Error(err))
				actionPair.Input.Error = "some error has ocurred"
			}
			output := &actionPair.Input
			output.ActionDone = true

			actionPair.ReplyTo <- *output

		default:
		}
	}
}

func RunWorker(ctx context.Context, config config.Worker, storageService StorageService) {
	workerLog := logger.Logger.With(zap.String("context", "worker"))

	actionCh := make(chan core.ActionReplyPair)
	var producerWg sync.WaitGroup
	var consumerWg sync.WaitGroup

	worker := &Worker{storage: storageService}

	producerWg.Add(1)
	go func() {
		worker.Produce(ctx, actionCh)
		producerWg.Done()
	}()

	consumerWg.Add(config.Number)
	for i := 0; i < config.Number; i++ {
		func() {
			go worker.Consume(ctx, actionCh)
			consumerWg.Done()
		}()
	}

	producerWg.Wait()
	workerLog.Debug("Closing producer")
	close(actionCh)

	consumerWg.Wait()
	workerLog.Debug("Closing consumers")
}
