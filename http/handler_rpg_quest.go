package http

import (
	"net/http"

	"github.com/SQUASHD/hbit/rpg/quest"
)

type questHandler struct {
	questSvc quest.UserQuestService
}

func newQuestHandler(questSvc quest.UserQuestService) *questHandler {
	return &questHandler{questSvc: questSvc}
}

func (h *questHandler) GetAll(w http.ResponseWriter, r *http.Request, userId string) {
	quests, err := h.questSvc.ListQuests(r.Context(), userId)
	if err != nil {
		Error(w, r, err)
		return
	}

	respondWithJSON(w, http.StatusOK, quests)
}
