package notion

import (
	"github.com/google/uuid"

	"oh-my-chat/src/service"
)

type StudyInspectCmd struct {
	RoadmapID string
}

func (c StudyInspectCmd) Meta() service.MessageMeta {
	return service.MessageMeta{
		Id:    uuid.New(),
		Topic: "notion_inspect_study_road_map",
	}
}
