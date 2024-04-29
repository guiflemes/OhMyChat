package main

import (
	"context"
	"log"
	"sync"

	"github.com/joho/godotenv"

	"oh-my-chat/src/core"
	"oh-my-chat/src/models"
	"oh-my-chat/src/schemas"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	schemas.ReadYml()

}

type mockWorkflowGetter struct {
	//guidedRepo GuidedRepo
}

func (m *mockWorkflowGetter) GetFlow(channelName string) core.Workflow {
	return &mockWorkflow{}
}

type mockWorkflow struct{}

func (m *mockWorkflow) Engine() string {
	return "guided"
}

func Run() {

	var (
		inputMsg  = make(chan models.Message, 1)
		outputMsg = make(chan models.Message, 1)
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot := models.NewBot(models.Telegram)
	actionQueue := core.NewGoActionQueue()
	actionQueue.Consume(ctx)
	guidedEngine := core.NewGuidedResponseEngine(actionQueue)

	processor := core.NewProcessor(&mockWorkflowGetter{}, core.Engines{guidedEngine})
	connector := core.NewMuitiChannelConnector(bot)
	var wg sync.WaitGroup

	wg.Add(3)

	go processor.Process(inputMsg, outputMsg)
	go connector.Request(inputMsg)
	go connector.Respose(outputMsg)

	wg.Wait()
}
