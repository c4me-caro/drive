package driver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/c4me-caro/drive"
	"github.com/c4me-caro/drive/cmd/auth"
	"github.com/c4me-caro/drive/database"
	"github.com/google/uuid"
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
	router.HandleFunc("/f/{file}", h.handleFile).Methods("GET")
	router.HandleFunc("/r/{file}", h.handleDeleteFile).Methods("GET")
	router.HandleFunc("/d/{folder}", h.handleFolder).Methods("GET")
	router.HandleFunc("/upload", h.handleUpload).Methods("POST")
	router.HandleFunc("/create", h.handleNewFolder).Methods("POST")
}

func (h Handler) handleFile(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	if file == "" {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: No file specified")
		return
	}

	// check for access permissions
	Authorization := r.Header.Get("Authorization")
	if Authorization == "" {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized")
		return
	}

	userId, err := auth.GetUserIdFromToken(Authorization)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized")
		return
	}

	user, err := h.db.FindUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding user")
		return
	}

	resource, err := h.db.GetResource("drives")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding resource")
		return
	}

	permissions := auth.FindPermission(user, "read", resource)
	if permissions == "" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Forbidden")
		return
	}

	// Check if file exists
	resource, err = h.db.GetResourceId(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding file")
		return
	}

	permissions = auth.FindPermission(user, "read", resource)
	if permissions == "" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Forbidden")
		return
	}

	filePath := "files/" + resource.Location
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error reading file")
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+resource.Name)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileBytes)))
	w.Write(fileBytes)
}

func (h Handler) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	file := mux.Vars(r)["file"]
	if file == "" {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: No file specified")
		return
	}

	// check for access permissions
	Authorization := r.Header.Get("Authorization")
	if Authorization == "" {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized")
		return
	}

	userId, err := auth.GetUserIdFromToken(Authorization)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized")
		return
	}

	user, err := h.db.FindUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding user")
		return
	}

	resource, err := h.db.GetResource("drives")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding resource")
		return
	}

	permissions := auth.FindPermission(user, "update", resource)
	if permissions == "" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Forbidden")
		return
	}

	// Check if file exists
	resource, err = h.db.GetResourceId(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding file")
		return
	}

	permissions = auth.FindPermission(user, "delete", resource)
	if permissions == "" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Forbidden")
		return
	}

	if resource.Type == "folder" && len(resource.Content) != 0 {
		w.WriteHeader(http.StatusConflict)
		io.WriteString(w, "Error: Folder not empty")
		return
	}

	err = h.db.DeleteResource(resource)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error trying to remove file")
		return
	}

	err = h.db.RemoveFileResource(resource.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error trying to remove file")
		return
	}

	io.WriteString(w, "File removed")
}

func (h Handler) handleUpload(w http.ResponseWriter, r *http.Request) {
	// check for access permissions
	Authorization := r.Header.Get("Authorization")
	if Authorization == "" {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized")
		return
	}

	userId, err := auth.GetUserIdFromToken(Authorization)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized")
		return
	}

	user, err := h.db.FindUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding user")
		return
	}

	resource, err := h.db.GetResource("drives")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding resource")
		return
	}

	permissions := auth.FindPermission(user, "create", resource)
	if permissions == "" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Forbidden")
		return
	}

	newUUID := uuid.New().String()
	body := drive.Resource{}

	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: File not found")
		return
	}

	parent := r.FormValue("parent")
	if parent != "" {
		resource, err = h.db.GetResourceId(parent)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Error: Finding parent resource")
			return
		}

		err = h.db.UpdateFolderContent(resource, newUUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Error: Updating parent resource")
			return
		}
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: Reading file")
		return
	}

	fileName := fmt.Sprintf("%s_%s", newUUID, handler.Filename)

	err = os.WriteFile("files/"+fileName, fileBytes, 0644)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: Writing file")
		return
	}

	body.Id = newUUID
	body.Name = handler.Filename
	body.OwnerId = userId
	body.SharedId = []string{}
	body.Type = "file"
	body.Location = fileName

	h.db.CreateResource(body)
	io.WriteString(w, body.Id)
}

func (h Handler) handleFolder(w http.ResponseWriter, r *http.Request) {
	folder := mux.Vars(r)["folder"]
	if folder == "" {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: No file specified")
		return
	}

	// check for access permissions
	Authorization := r.Header.Get("Authorization")
	if Authorization == "" {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized")
		return
	}

	userId, err := auth.GetUserIdFromToken(Authorization)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized")
		return
	}

	user, err := h.db.FindUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding user")
		return
	}

	resource, err := h.db.GetResource("drives")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding resource")
		return
	}

	permissions := auth.FindPermission(user, "read", resource)
	if permissions == "" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Forbidden")
		return
	}

	// Check if folder exists
	resource, err = h.db.GetResourceId(folder)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding file")
		return
	}

	permissions = auth.FindPermission(user, "read", resource)
	if permissions == "" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Forbidden")
		return
	}

	if resource.Type == "file" {
		w.WriteHeader(http.StatusConflict)
		io.WriteString(w, "Resource is a file")
		return
	}

	if err := json.NewEncoder(w).Encode(resource.Content); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error coding response")
		return
	}

}

func (h Handler) handleNewFolder(w http.ResponseWriter, r *http.Request) {
	// check for access permissions
	Authorization := r.Header.Get("Authorization")
	if Authorization == "" {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized")
		return
	}

	userId, err := auth.GetUserIdFromToken(Authorization)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized")
		return
	}

	user, err := h.db.FindUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding user")
		return
	}

	resource, err := h.db.GetResource("drives")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error finding resource")
		return
	}

	permissions := auth.FindPermission(user, "create", resource)
	if permissions == "" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Forbidden")
		return
	}

	name := r.FormValue("name")
	if name == "" {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: No folder name specified")
		return
	}

	newUUID := uuid.New().String()
	var body drive.Resource

	parent := r.FormValue("parent")
	if parent != "" {
		resource, err = h.db.GetResourceId(parent)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Error: Finding parent resource")
			return
		}

		err = h.db.UpdateFolderContent(resource, newUUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Error: Updating parent resource")
			return
		}
	}

	body.Id = newUUID
	body.Name = name
	body.OwnerId = userId
	body.SharedId = []string{}
	body.Location = ""
	body.Type = "folder"
	body.Content = []string{}

	h.db.CreateResource(body)
	io.WriteString(w, body.Id)
}
