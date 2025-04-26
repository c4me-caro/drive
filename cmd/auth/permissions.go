package auth

import (
	"sync"

	"github.com/c4me-caro/drive"
)

var permissionCache sync.Map

func FindPermission(user drive.User, access string, resource drive.Resource) string {
	if resource.Id == "0" {
		return handleSystemResource(user, access)
	}

	sharedPermission := access + ":" + user.Id + "-" + resource.Name
	if validatePermission(user, sharedPermission) && searchSharedId(user, resource) {
		return sharedPermission
	}

	candidates := make([]string, 0, 7)
	candidates = append(candidates, access+":"+resource.Name)
	candidates = append(candidates, access+":all")
	candidates = append(candidates, "all:"+resource.Name)

	if user.Id == resource.OwnerId {
		candidates = append(candidates, access+"own-"+resource.Name)
		candidates = append(candidates, access+"own-all")
		candidates = append(candidates, "all:own-all")
	}

	candidates = append(candidates, "all:all")

	for _, candidate := range candidates {
		if validatePermission(user, candidate) {
			return candidate
		}
	}

	return ""
}

func handleSystemResource(user drive.User, access string) string {
	if access == "read" {
		if validatePermission(user, "read:sys-all") {
			return "read:sys-all"
		}

		if validatePermission(user, "all:sys-all") {
			return "all:sys-all"
		}
	}

	if user.Role == "admin" && (access == "create" || access == "update") {
		return access + ":sys-all"
	}

	return ""
}

func validatePermission(user drive.User, access string) bool {
	if cached, ok := permissionCache.Load(user.Id); ok {
		permSet := cached.(map[string]struct{})
		_, exists := permSet[access]
		return exists
	}

	permSet := make(map[string]struct{})
	for _, p := range user.Permissions {
		permSet[p] = struct{}{}
	}
	permissionCache.Store(user.Id, permSet)

	_, exists := permSet[access]
	return exists
}

func searchSharedId(user drive.User, resource drive.Resource) bool {
	for _, id := range resource.SharedId {
		if id == user.Id {
			return true
		}
	}

	return false
}
