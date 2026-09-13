package httpapi

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tariq-ventura/entropy-mcp-server/internal/integration"
	"github.com/tariq-ventura/entropy-mcp-server/internal/projection"
)

func New(service *integration.Service, apiKey, allowedOrigin string) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), cors(allowedOrigin), bearer(apiKey))
	api := router.Group("/api/v1")
	api.GET("/dashboard", func(c *gin.Context) {
		data, err := service.Dashboard(c.Request.Context())
		respond(c, data, err)
	})
	api.GET("/equipment-types", func(c *gin.Context) {
		data, err := service.ListTypes(c.Request.Context())
		respond(c, data, err)
	})
	api.GET("/equipments", func(c *gin.Context) {
		page, pageSize := pagination(c)
		data, total, err := service.ListEquipment(c.Request.Context(), page, pageSize, c.Query("type"), c.Query("status"), c.Query("search"), boolQuery(c, "onlyAvailable"), boolQuery(c, "onlyLinked"))
		respondList(c, data, page, pageSize, total, err)
	})
	api.GET("/equipments/:key", func(c *gin.Context) {
		data, err := service.GetEquipment(c.Request.Context(), c.Param("key"))
		respond(c, data, err)
	})
	api.GET("/requests", func(c *gin.Context) {
		page, pageSize := pagination(c)
		data, total, err := service.ListRequests(c.Request.Context(), page, pageSize, c.Query("search"), c.Query("status"), c.Query("type"), c.Query("requester"))
		respondList(c, data, page, pageSize, total, err)
	})
	api.GET("/assignments", func(c *gin.Context) {
		page, pageSize := pagination(c)
		data, total, err := service.ListAssignments(c.Request.Context(), page, pageSize, c.Query("status"), c.Query("search"))
		respondList(c, data, page, pageSize, total, err)
	})
	api.GET("/conflicts", func(c *gin.Context) {
		page, pageSize := pagination(c)
		data, total, err := service.ListConflicts(c.Request.Context(), page, pageSize)
		respondList(c, data, page, pageSize, total, err)
	})
	api.GET("/sync-status", func(c *gin.Context) {
		data, err := service.GetSyncState(c.Request.Context())
		respond(c, data, err)
	})
	api.POST("/sync", func(c *gin.Context) {
		if err := service.Sync(c.Request.Context()); err != nil {
			respond(c, nil, err)
			return
		}
		data, err := service.GetSyncState(c.Request.Context())
		respond(c, data, err)
	})
	return router
}

func bearer(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		provided := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(c.GetHeader("Authorization")), "Bearer "))
		if len(provided) != len(expected) || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Bearer token inválido"})
			return
		}
		c.Next()
	}
}

func cors(origin string) gin.HandlerFunc {
	origin = strings.TrimSpace(origin)
	return func(c *gin.Context) {
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		c.Next()
	}
}

func pagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func boolQuery(c *gin.Context, name string) bool {
	value, _ := strconv.ParseBool(c.Query(name))
	return value
}

func respond(c *gin.Context, data any, err error) {
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, projection.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": http.StatusText(status), "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func respondList(c *gin.Context, data any, page, pageSize int, total int64, err error) {
	if err != nil {
		respond(c, nil, err)
		return
	}
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "pagination": gin.H{"page": page, "pageSize": pageSize, "total": total, "totalPages": totalPages}})
}
