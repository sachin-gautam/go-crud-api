package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/sachin-gautam/go-crud-api/internal/model"
	"github.com/sachin-gautam/go-crud-api/internal/storage"
	"github.com/sachin-gautam/go-crud-api/internal/utils/response"
)

type StudentHandler struct {
	storage storage.Storage
}

func NewStudentHandler(storage storage.Storage) StudentHandler {
	return StudentHandler{storage}
}

func (h StudentHandler) Create(w http.ResponseWriter, r *http.Request) {
	slog.Info("Creating a Student")

	var student model.Student

	err := json.NewDecoder(r.Body).Decode(&student)

	if errors.Is(err, io.EOF) {
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
		return
	}

	if err != nil {
		response.WriteJson(w, http.StatusBadGateway, response.GeneralError(err))
		return
	}

	//Request Validation
	if err := validator.New().Struct(student); err != nil {
		validateErrs := err.(validator.ValidationErrors)
		response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
		return
	}

	lastId, err := h.storage.CreateStudent(
		student.Name,
		student.Email,
		student.Age,
	)

	slog.Info("User Created Successfully", slog.String("userId", fmt.Sprint(lastId)))

	if err != nil {
		response.WriteJson(w, http.StatusInternalServerError, err)
		return
	}

	response.WriteJson(w, http.StatusCreated, map[string]int64{"id": lastId})
}

func (h StudentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	slog.Info("getting a student", slog.String("id", id))

	intId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}
	student, err := h.storage.GetStudentById(intId)
	if err != nil {
		slog.Error("Error getting user", slog.String("id", id))
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}
	response.WriteJson(w, http.StatusOK, student)
}

func (h StudentHandler) GetList(w http.ResponseWriter, r *http.Request) {
	slog.Info("getting all students")
	students, err := h.storage.GetList()
	if err != nil {
		response.WriteJson(w, http.StatusInternalServerError, err)
		return
	}
	response.WriteJson(w, http.StatusOK, students)
}

func (h StudentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	slog.Info("updating student", slog.String("id", id))

	intId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("invalid id format: %w", err)))
		return
	}

	var student model.Student
	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("invalid request payload: %w", err)))
		return
	}

	if err := validator.New().Struct(student); err != nil {
		validateErrs := err.(validator.ValidationErrors)
		response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
		return
	}

	updatedStudent, err := h.storage.UpdateById(intId, student.Name, student.Email, student.Age)
	if err != nil {
		slog.Error("Error updating student", slog.String("id", id), slog.Any("error", err))
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	response.WriteJson(w, http.StatusOK, updatedStudent)
}

func (h StudentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	slog.Info("deleting a student", slog.String("id", id))

	intId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("invalid student ID")))
		return
	}

	deletedId, err := h.storage.DeleteById(intId)
	if err != nil {
		slog.Error("error deleting student", slog.String("id", id))
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	response.WriteJson(w, http.StatusOK, map[string]int64{"deleted_id": deletedId})
}
