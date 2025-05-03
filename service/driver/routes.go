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
	"github.com/joho/godotenv"
)

type Handler struct {
	db *database.DriveWorker
}

func NewHandler(db *database.DriveWorker) *Handler {
	godotenv.Load()
	return &Handler{
		db: db,
	}
}

func (h Handler) checkResource(resname string, user drive.User, operation string) (drive.Resource, error) {
	resource, err := h.db.GetResource(resname)
	if err != nil {
		return drive.Resource{}, err
	}

	permissions := auth.FindPermission(user, operation, resource)
	if permissions == "" {
		return drive.Resource{}, fmt.Errorf("user permission property not found")
	}

	return resource, nil
}

func (h Handler) validateAuthentication(r *http.Request, operation string) (drive.User, error) {
	Authorization := r.Header.Get("Authorization")
	if Authorization == "" {
		return drive.User{}, fmt.Errorf("invalid token: %s", Authorization)
	}

	userId, err := auth.GetUserIdFromToken(Authorization)
	if err != nil {
		return drive.User{}, err
	}

	user, err := h.db.GetUserById(userId)
	if err != nil {
		return drive.User{}, err
	}

	_, err = h.checkResource("drive", user, operation)
	if err != nil {
		return drive.User{}, err
	}

	return user, nil
}

func (h Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/f/{file}", h.handleFile).Methods("GET")
	router.HandleFunc("/d/{folder}", h.handleFolder).Methods("GET")
	router.HandleFunc("/r/{file}", h.handleDeleteFile).Methods("GET")
	router.HandleFunc("/rd/{folder}", h.handleDeleteFolder).Methods("GET")
	router.HandleFunc("/create/{folder}", h.handleNewFolder).Methods("POST")
	router.HandleFunc("/upload/{parent}", h.handleNewFile).Methods("POST")
}

func (h Handler) handleFile(w http.ResponseWriter, r *http.Request) {
	user, err := h.validateAuthentication(r, "read")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Unauthorized")
		return
	}

	file := mux.Vars(r)["file"]
	if file == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Error: No file specified")
		return
	}

	resource, err := h.checkResource(file, user, "read")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Unauthorized")
		return
	}

	filePath := resource.Location
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: file not found")
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+resource.Name)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileBytes)))
	w.Write(fileBytes)
}

func (h Handler) handleFolder(w http.ResponseWriter, r *http.Request) {
	user, err := h.validateAuthentication(r, "read")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Unauthorized")
		return
	}

	folder := mux.Vars(r)["folder"]
	if folder == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Error: No folder specified")
		return
	}

	resource, err := h.checkResource(folder, user, "read")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Unauthorized")
		return
	}

	if resource.Type != "folder" {
		w.WriteHeader(http.StatusConflict)
		io.WriteString(w, "Error: resource is not a folder")
		return
	}

	if err := json.NewEncoder(w).Encode(resource); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: bad json coding")
		return
	}
}

func (h Handler) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	user, err := h.validateAuthentication(r, "delete")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Unauthorized")
		return
	}

	file := mux.Vars(r)["file"]
	if file == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Error: No file specified")
		return
	}

	resource, err := h.checkResource(file, user, "delete")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Unauthorized")
		return
	}

	if resource.Type != "file" {
		w.WriteHeader(http.StatusConflict)
		io.WriteString(w, "Error: Resource is not a file")
		return
	}

	err = h.db.DeleteResource(resource)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: resource cannot be deleted")
		return
	}

	io.WriteString(w, "File removed")
}

func (h Handler) handleDeleteFolder(w http.ResponseWriter, r *http.Request) {
	user, err := h.validateAuthentication(r, "delete")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Unauthorized")
		return
	}

	folder := mux.Vars(r)["folder"]
	if folder == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Error: No folder specified")
		return
	}

	resource, err := h.checkResource(folder, user, "delete")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Unauthorized")
		return
	}

	if resource.Type != "folder" {
		w.WriteHeader(http.StatusConflict)
		io.WriteString(w, "Error: Resource is not a folder")
		return
	}

	recursive := r.URL.Query().Get("recursive")
	childrens := len(resource.Content)

	if childrens != 0 && recursive == "false" {
		w.WriteHeader(http.StatusConflict)
		io.WriteString(w, "Error: Folder not empty")
	}

	deletionCounter := 0
	if recursive == "true" {
		for _, rsx := range resource.Content {

			res, err := h.checkResource(rsx, user, "delete")
			if err != nil {
				deletionCounter++
				continue
			}

			err = h.db.DeleteResource(res)
			if err != nil {
				deletionCounter++
				continue
			}
		}
	}

	err = h.db.DeleteResource(resource)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprintf("Error: resource cannot be deleted. Deleted children: %d", deletionCounter))
		return
	}

	io.WriteString(w, fmt.Sprintf("Folder removed. Deleted children: %d", deletionCounter))
}

func (h Handler) handleNewFolder(w http.ResponseWriter, r *http.Request) {
	user, err := h.validateAuthentication(r, "create")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Unauthorized")
		return
	}

	folder := mux.Vars(r)["folder"]
	if folder == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Error: No folder specified")
		return
	}

	reqBody, _ := io.ReadAll(r.Body)
	var parent drive.Resource
	json.Unmarshal(reqBody, &parent)
	newUUID := uuid.New().String()
	container := ""

	if parent.Name != "" {
		err = h.db.CheckResource(parent)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "Error: Unauthorized")
			return
		}

		permissions := auth.FindPermission(user, "update", parent)
		if permissions == "" {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "Error: Unauthorized")
			return
		}

		err = h.db.AddResourceChildren(parent, newUUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Error: Failed parent update")
			return
		}

		container = parent.Name
	}

	var body drive.Resource

	body.Id = newUUID
	body.Name = folder
	body.OwnerId = user.Id
	body.SharedId = []string{}
	body.Location = container
	body.Type = "folder"
	body.Content = []string{}

	err = h.db.CreateResource(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: Failed resource creation")
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: bad json coding")
		return
	}
}

func (h Handler) handleNewFile(w http.ResponseWriter, r *http.Request) {
	user, err := h.validateAuthentication(r, "create")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Error: Unauthorized")
		return
	}

	newUUID := uuid.New().String()
	parent := mux.Vars(r)["parent"]
	if parent != "" {
		resource, err := h.checkResource(parent, user, "update")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "Error: Unauthorized")
			return
		}

		err = h.db.AddResourceChildren(resource, newUUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Error: Failed parent update")
			return
		}
	}

	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: File not found")
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: Reading file")
		return
	}

	fileName := fmt.Sprintf("%s_%s", newUUID, handler.Filename)
	err = os.WriteFile(os.Getenv("FILES_ROOT")+"/"+fileName, fileBytes, 0644)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: Writing file")
		return
	}

	var body drive.Resource

	body.Id = newUUID
	body.Name = fileName
	body.OwnerId = user.Id
	body.SharedId = []string{}
	body.Location = os.Getenv("FILES_ROOT") + "/" + fileName
	body.Type = "file"
	body.Content = []string{}

	err = h.db.CreateResource(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: Failed resource creation")
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Error: bad json coding")
		return
	}
}
