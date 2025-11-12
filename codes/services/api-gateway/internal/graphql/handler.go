package graphqlhandler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/rs/zerolog"

	"github.com/venue-master/platform/lib/errutil"
	"github.com/venue-master/platform/lib/jwtutil"
	"github.com/venue-master/platform/services/api-gateway/internal/services"
)

type claimsKey struct{}

// Handler wires HTTP traffic into the GraphQL schema.
type Handler struct {
	schema  graphql.Schema
	jwt     *jwtutil.Manager
	logger  zerolog.Logger
	clients *services.ServiceClients
}

// GraphQLRequest matches the standard GraphQL over HTTP contract.
type GraphQLRequest struct {
	OperationName string                 `json:"operationName"`
	Query         string                 `json:"query" binding:"required"`
	Variables     map[string]interface{} `json:"variables"`
}

// New builds a Handler with schema + dependencies.
func New(clients *services.ServiceClients, jwt *jwtutil.Manager, logger zerolog.Logger) (*Handler, error) {
	schema, err := buildSchema(clients)
	if err != nil {
		return nil, err
	}

	return &Handler{schema: schema, jwt: jwt, logger: logger, clients: clients}, nil
}

// Register attaches the GraphQL endpoints to the gin engine.
func (h *Handler) Register(router *gin.Engine) {
	router.POST("/graphql", h.handleGraphQL)
	router.GET("/graphql", h.handleGraphQL)
}

func (h *Handler) handleGraphQL(ctx *gin.Context) {
	var req GraphQLRequest
	if ctx.Request.Method == http.MethodGet {
		req.Query = ctx.Query("query")
		req.OperationName = ctx.Query("operationName")
	} else if err := ctx.ShouldBindJSON(&req); err != nil {
		errutil.Write(ctx, http.StatusBadRequest, "invalid_request", "Invalid GraphQL payload", err.Error())
		return
	}

	if strings.TrimSpace(req.Query) == "" {
		errutil.Write(ctx, http.StatusBadRequest, "missing_query", "Query cannot be empty", nil)
		return
	}

	claims := h.extractClaims(ctx.Request)
	requestCtx := context.WithValue(ctx.Request.Context(), claimsKey{}, claims)
	if claims != nil {
		requestCtx = services.WithAuth(requestCtx, services.AuthMetadata{
			UserID: claims.UserID,
			Roles:  claims.Roles,
		})
	}

	result := graphql.Do(graphql.Params{
		Schema:         h.schema,
		RequestString:  req.Query,
		OperationName:  req.OperationName,
		VariableValues: req.Variables,
		Context:        requestCtx,
	})

	status := http.StatusOK
	if len(result.Errors) > 0 {
		status = http.StatusBadRequest
	}
	ctx.JSON(status, result)
}

// ClaimsFromContext extracts JWT claims for resolvers.
func ClaimsFromContext(ctx context.Context) *jwtutil.Claims {
	claims, _ := ctx.Value(claimsKey{}).(*jwtutil.Claims)
	return claims
}

func (h *Handler) extractClaims(r *http.Request) *jwtutil.Claims {
	defaultClaims := &jwtutil.Claims{UserID: "user-1", Roles: []string{"MEMBER"}}

	authz := r.Header.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return defaultClaims
	}
	token := strings.TrimSpace(authz[7:])
	claims, err := h.jwt.Validate(token)
	if err != nil {
		h.logger.Warn().Err(err).Msg("jwt validation failed, falling back to default claims")
		return defaultClaims
	}
	return claims
}
