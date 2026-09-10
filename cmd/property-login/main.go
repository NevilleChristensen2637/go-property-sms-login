package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/example/property-sms-login/internal/infrai"
	"github.com/example/property-sms-login/internal/property"
)

type loginRequest struct {
	Phone     string `json:"phone"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id"`
}

type server struct {
	login *property.LoginService
}

func main() {
	client, err := infrai.NewClient(os.Getenv("INFRAI_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	s := &server{login: property.NewLoginService(client)}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login/code", s.requestCode)
	mux.HandleFunc("POST /login/verify", s.verifyCode)
	log.Println("property login listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (s *server) requestCode(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if err := decode(r, &input); err != nil || input.Phone == "" || input.RequestID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "phone and request_id are required"})
		return
	}
	if err := s.login.RequestCode(r.Context(), input.Phone, input.RequestID); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "code_sent", "phone": maskPhone(input.Phone)})
}

func (s *server) verifyCode(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if err := decode(r, &input); err != nil || input.Phone == "" || input.Code == "" || input.RequestID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "phone, code, and request_id are required"})
		return
	}
	dashboard, err := s.login.VerifyCode(r.Context(), input.Phone, input.Code, input.RequestID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dashboard)
}

func decode(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, property.ErrUnknownTenant) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	var apiErr *infrai.APIError
	if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
		writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Message})
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": "login provider request failed"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}

func maskPhone(phone string) string {
	if len(phone) < 4 {
		return strings.Repeat("*", len(phone))
	}
	return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}
