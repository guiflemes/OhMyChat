package core_test

import (
	"testing"
	"time"

	"oh-my-chat/src/core"
	"oh-my-chat/src/core/mocks"
	"oh-my-chat/src/message"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestProcessor_Process(t *testing.T) {
	t.Parallel()

	t.Run("sucessfully", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockEngine := mocks.NewMockEngine(ctrl)

		mockEngine.EXPECT().HandleMessage(gomock.Any(), gomock.Any()).Times(1)

		proc := core.NewProcessor(mockEngine)

		input := make(chan message.Message, 1)
		output := make(chan message.Message, 1)

		msg := message.Message{
			User: message.User{ID: "user123"},
		}

		ctx := core.NewChatContext()

		input <- msg

		go proc.Process(ctx, input, output)

		time.Sleep(100 * time.Millisecond)
		ctx.Shutdown()

	})

	t.Run("skip handling message, session adapter error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockEngine := mocks.NewMockEngine(ctrl)
		mockSessionAdapter := mocks.NewMockSessionAdapter(ctrl)

		mockEngine.EXPECT().HandleMessage(gomock.Any(), gomock.Any()).Times(0)
		mockSessionAdapter.EXPECT().GetOrCreate(gomock.Any(), "user123").Return(nil, assert.AnError).Times(1)

		proc := core.NewProcessor(mockEngine)

		input := make(chan message.Message, 1)
		output := make(chan message.Message, 1)

		msg := message.Message{
			User: message.User{ID: "user123"},
		}

		ctx := core.NewChatContext(core.WithSessionAdapter(mockSessionAdapter))

		input <- msg

		go proc.Process(ctx, input, output)

		time.Sleep(100 * time.Millisecond)
		ctx.Shutdown()

	})
}
