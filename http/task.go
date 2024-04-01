package http

import (
	"net/http"

	"github.com/SQUASHD/hbit/task"
)

func (s *ServerMonolith) handleTaskFindAll(w http.ResponseWriter, r *http.Request, userId string) {
	tasks, err := s.taskSvc.List(r.Context(), userId)
	if err != nil {
		Error(w, r, err)
		return
	}
	RespondWithJSON(w, http.StatusOK, tasks)
}

func (h *ServerMonolith) Get(w http.ResponseWriter, r *http.Request, requestedById string) {

	todos, err := h.taskSvc.List(r.Context(), requestedById)
	if err != nil {
		Error(w, r, err)
		return
	}
	RespondWithJSON(w, http.StatusOK, todos)

}

func (h *ServerMonolith) handleTaskCreate(w http.ResponseWriter, r *http.Request, requestedById string) {
	var data task.CreateTaskData

	if err := Decode(r, &data); err != nil {
		Error(w, r, err)
		return
	}

	form := task.CreateTaskForm{
		CreateTaskData: data,
		RequestedById:  requestedById,
	}

	todo, err := h.taskSvc.Create(r.Context(), form)
	if err != nil {
		Error(w, r, err)
		return
	}

	RespondWithJSON(w, http.StatusCreated, todo)
}

func (h *ServerMonolith) handleTaskUpdate(w http.ResponseWriter, r *http.Request, requestedById string) {
	id := r.PathValue("id")
	var data task.UpdateTaskData

	if err := Decode(r, &data); err != nil {
		Error(w, r, err)
		return
	}

	form := task.UpdateTaskForm{
		UpdateTaskData: data,
		TaskId:         id,
		RequestedById:  requestedById,
	}

	todo, err := h.taskSvc.Update(r.Context(), form)
	if err != nil {
		Error(w, r, err)
		return
	}

	RespondWithJSON(w, http.StatusOK, todo)
}

func (h *ServerMonolith) handleTaskDelete(w http.ResponseWriter, r *http.Request, requestedById string) {
	id := r.PathValue("id")

	form := task.DeleteTaskForm{
		TaskId:        id,
		RequestedById: requestedById,
	}

	err := h.taskSvc.Delete(r.Context(), form)
	if err != nil {
		Error(w, r, err)
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
