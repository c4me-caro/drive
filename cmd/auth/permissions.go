package auth

import (
	"strings"

	"github.com/c4me-caro/drive"
)

func FindPermission(user drive.User, access string, resource drive.Resource) string {
	if user.Role == "admin" && resource.OwnerId == "0" && access != "update" && access != "delete" {
		return access + ":sys-all"
	}

	sharedPermission := access + ":" + resource.Id + "-" + resource.Name
	if validatePermission(user, sharedPermission) && searchSharedId(user, resource) {
		return sharedPermission
	}

	candidates := []string{
		access + ":" + resource.Name,
		access + ":all",
		"all:" + resource.Name,
		"all:all",
	}

	if user.Id == resource.OwnerId {
		candidates = append(candidates, access+":own-"+resource.Name)
		candidates = append(candidates, access+":own-all")
		candidates = append(candidates, "all:own-all")
	}

	for _, candidate := range candidates {
		if validatePermission(user, candidate) {
			return candidate
		}
	}

	return ""
}

func validatePermission(user drive.User, access string) bool {
	for _, permission := range user.Permissions {
		if strings.EqualFold(permission, access) {
			return true
		}
	}

	return false
}

func searchSharedId(user drive.User, resource drive.Resource) bool {
	for _, id := range resource.SharedId {
		if id == user.Id {
			return true
		}
	}

	return false
}
