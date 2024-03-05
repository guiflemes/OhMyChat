package notion

import (
	"log"
	"notion-agenda/service"
)

type RoadMapGetter interface {
	GetRoadMap(pageID string) (*Roadmap, error)
}

func pendency(roadMap *Roadmap, publisher Publisher) {
	if roadMap.HasPendency() {
		publisher.Publish(&PendencyEvent{Pendency: roadMap.Pendency()})
	}
}

type studyInspectHandler struct {
	notionRepo  RoadMapGetter
	inspections []func(roadMap *Roadmap, publisher Publisher)
	publisher   Publisher
}

func NewStudyInspectHandler(notionRepo RoadMapGetter, publisher Publisher) *studyInspectHandler {
	return &studyInspectHandler{
		notionRepo: notionRepo,
		publisher:  publisher,
		inspections: []func(roadMap *Roadmap, publisher Publisher){
			pendency,
		},
	}
}

func (h *studyInspectHandler) Handle(message service.Message) error {
	cmd, ok := message.(*StudyInspectCmd)

	if !ok {
		log.Printf("Unexpected type in Function: %T", message)
		panic("Critical error: Unexpected type")
	}

	roadMap, err := h.notionRepo.GetRoadMap(cmd.RoadmapID)

	if err != nil {
		return err
	}

	for _, insp := range h.inspections {
		insp(roadMap, h.publisher)
	}

	return nil
}