package user

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/c4me-caro/drive/cmd/auth"
	"github.com/c4me-caro/drive/database"
	"github.com/gorilla/mux"
)

type Handler struct {
	db *database.DriveWorker
}

func NewHandler(db *database.DriveWorker) *Handler {
	return &Handler{
		db: db,
	}
}

func (h Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/validateUser", h.handleValidUser).Methods("GET")
	router.HandleFunc("/logout", h.handleLogout).Methods("GET")
}

func (h Handler) handleValidUser(w http.ResponseWriter, r *http.Request) {
	Authorization := r.Header.Get("Authorization")
	if Authorization == "" {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Unauthorized")
		return
	}

	userId, err := auth.GetUserIdFromToken(Authorization)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Unauthorized")
		return
	}

	_, err = h.db.GetUserById(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: user not found")
		return
	}

	io.WriteString(w, Authorization)
}

func (h Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	type test_struct struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	reqBody, _ := io.ReadAll(r.Body)
	var body test_struct
	json.Unmarshal(reqBody, &body)

	// validate login credentials
	user, err := h.db.GetUser(body.Username, body.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Username or Password is incorrect")
		return
	}

	// generate access token
	token, err := auth.CreateJWT(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: token generation failed")
		return
	}

	io.WriteString(w, token)
}

func (h Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	// invalidate token
	auth.InvalidateToken(r.Header.Get("Authorization"))
	io.WriteString(w, "Logout successfully")
}
