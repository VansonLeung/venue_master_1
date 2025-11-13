package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const contextUserKey = "authUser"

const (
	RoleAdmin      = "ADMIN"
	RoleVenueAdmin = "VENUE_ADMIN"
	RoleOperator   = "OPERATOR"
	RoleMember     = "MEMBER"
)

// ContextUser represents the authenticated caller.
type ContextUser struct {
	UserID string
	Roles  []string
}

// HasRole checks for a role.
func (u ContextUser) HasRole(role string) bool {
	role = strings.ToUpper(role)
	for _, r := range u.Roles {
		if strings.ToUpper(strings.TrimSpace(r)) == role {
			return true
		}
	}
	return false
}

// HasAnyRole returns true if the user owns any allowed role.
func (u ContextUser) HasAnyRole(roles ...string) bool {
	for _, target := range roles {
		if target != "" && u.HasRole(target) {
			return true
		}
	}
	return false
}

// RequireAuth validates headers and stores context user.
func RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := strings.TrimSpace(ctx.GetHeader("X-User-ID"))
		rolesHeader := strings.TrimSpace(ctx.GetHeader("X-User-Roles"))
		if userID == "" || rolesHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth headers"})
			return
		}
		roles := []string{}
		for _, part := range strings.Split(rolesHeader, ",") {
			trimmed := strings.ToUpper(strings.TrimSpace(part))
			if trimmed != "" {
				roles = append(roles, trimmed)
			}
		}
		if len(roles) == 0 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing roles"})
			return
		}
		ctx.Set(contextUserKey, ContextUser{UserID: userID, Roles: roles})
		ctx.Next()
	}
}

// RequireRoles enforces allowed roles.
func RequireRoles(allowed ...string) gin.HandlerFunc {
	normalized := make([]string, 0, len(allowed))
	for _, r := range allowed {
		normalized = append(normalized, strings.ToUpper(strings.TrimSpace(r)))
	}
	return func(ctx *gin.Context) {
		user, ok := GetUser(ctx)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if !user.HasAnyRole(normalized...) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		ctx.Next()
	}
}

// GetUser retrieves auth context.
func GetUser(ctx *gin.Context) (ContextUser, bool) {
	val, ok := ctx.Get(contextUserKey)
	if !ok {
		return ContextUser{}, false
	}
	user, ok := val.(ContextUser)
	return user, ok
}
