package external

import "oh-my-chat/src/models"

type External interface {
	Acquire(input chan<- models.Message)
	Dispatch(message models.Message)
}
