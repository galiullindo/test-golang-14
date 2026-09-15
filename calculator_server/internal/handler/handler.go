package handler

import (
	"net/http"
	"strconv"

	"github.com/galiullindo/test-golang-14/calculator_server/internal/service"
)

type Handler struct {
	serv *service.Service
}

func NewHandler(serv *service.Service) *Handler {
	return &Handler{serv: serv}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /calc", h.PostCalc)
}

func (h *Handler) PostCalc(w http.ResponseWriter, r *http.Request) {
	strNum := r.URL.Query().Get("num")
	if strNum == "" {
		respond(w, http.StatusBadRequest, []byte("missing 'num' query parameter"))
		return
	}

	num, err := strconv.Atoi(strNum)
	if err != nil {
		respond(w, http.StatusBadRequest, []byte("'num' must be an integer"))
		return
	}

	h.serv.ProcessCalc(num)

	respond(w, http.StatusOK, []byte("ok"))
}

func respond(w http.ResponseWriter, code int, body []byte) {
	w.WriteHeader(code)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	if len(body) != 0 {
		w.Write(body)
	}
}
