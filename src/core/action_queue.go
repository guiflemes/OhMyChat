package core

import (
	"context"

	"go.uber.org/zap"

	"oh-my-chat/src/logger"
	"oh-my-chat/src/models"
)

type ActionReplyPair struct {
	replyTo chan<- models.Message
	action  Action
	input   models.Message
}

type goActionQueue struct {
	actionPair chan ActionReplyPair
	ctx        context.Context
}

func NewGoActionQueue() *goActionQueue {
	return &goActionQueue{}
}

func (q *goActionQueue) Put(ctx context.Context, actionPair ActionReplyPair) {
	q.actionPair <- actionPair
	q.ctx = ctx
}

func (q *goActionQueue) Consume() {

	go func() {
		for {
			select {
			case actionPair := <-q.actionPair:
				err := actionPair.action.Handle(q.ctx, &actionPair.input)

				if err != nil {
					logger.Logger.Error("Error Handling Action",
						zap.String("context", "goActionQueue"),
						zap.Error(err),
					)
				}

				actionPair.replyTo <- actionPair.input

			case <-q.ctx.Done():
				q.brodcastAll()
			}
		}
	}()
}

func (q *goActionQueue) brodcastAll() {
	for {
		select {
		case actionPair := <-q.actionPair:
			actionPair.input.Output = "Server is shutting down. Please reconnect later"

			logger.Logger.Warn("Shutting Down",
				zap.String("context", "goActionQueue"),
				zap.String("message", "context done"),
			)

			actionPair.replyTo <- actionPair.input
		default:
			return
		}
	}
}
