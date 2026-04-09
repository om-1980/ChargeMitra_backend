package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/om-1980/ChargeMitra_backend/configs"
	"github.com/om-1980/ChargeMitra_backend/internal/auth"
	"github.com/om-1980/ChargeMitra_backend/internal/chargers"
	"github.com/om-1980/ChargeMitra_backend/internal/middleware"
	"github.com/om-1980/ChargeMitra_backend/internal/stations"
	"github.com/om-1980/ChargeMitra_backend/internal/users"
	"github.com/om-1980/ChargeMitra_backend/pkg/response"
)

func Register(router *gin.Engine, db *pgxpool.Pool, cfg *configs.Config) {
	router.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "API is running", gin.H{
			"service": "api",
			"status":  "ok",
		})
	})

	authHandler := auth.NewHandler(db, cfg)
	userHandler := users.NewHandler(db)
	stationHandler := stations.NewHandler(db)
	chargerHandler := chargers.NewHandler(db)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			response.Success(c, http.StatusOK, "v1 API is running", gin.H{
				"version": "v1",
				"status":  "ok",
			})
		})

		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		protected := v1.Group("/me")
		protected.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			protected.GET("", func(c *gin.Context) {
				response.Success(c, http.StatusOK, "authenticated user", gin.H{
					"user_id":    c.GetString("user_id"),
					"user_email": c.GetString("user_email"),
					"user_role":  c.GetString("user_role"),
				})
			})

			protected.GET("/profile", userHandler.GetProfile)
		}

		stationGroup := v1.Group("/stations")
		stationGroup.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			stationGroup.GET("", stationHandler.List)
			stationGroup.GET("/my", stationHandler.MyStations)
			stationGroup.POST("", middleware.RequireRoles("owner", "admin"), stationHandler.Create)
		}

		chargerGroup := v1.Group("/chargers")
		chargerGroup.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			chargerGroup.GET("", chargerHandler.List)
			chargerGroup.GET("/station/:stationId", chargerHandler.ListByStation)
			chargerGroup.POST("", middleware.RequireRoles("owner", "admin"), chargerHandler.Create)
		}

		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuth(cfg.JWTSecret), middleware.RequireRoles("admin"))
		{
			admin.GET("/health", func(c *gin.Context) {
				response.Success(c, http.StatusOK, "admin route accessible", gin.H{
					"role": "admin",
				})
			})
		}
	}
}