package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterPetInternalRoutes registers server-to-server API endpoints
// used by the TAI Protocol backend to manage pet compute accounts.
// Protected by a shared secret header (X-Internal-Secret), not JWT.
func RegisterPetInternalRoutes(
	v1 *gin.RouterGroup,
	petHandler *handler.PetInternalHandler,
	internalSecret string,
) {
	pet := v1.Group("/internal/pet")
	pet.Use(middleware.InternalSecretGuard(internalSecret))
	{
		// Provision a new pet account (creates 3api user + API key)
		pet.POST("/provision", petHandler.Provision)

		// Credit compute balance to a pet (called when TAI is burned)
		pet.POST("/credit", petHandler.Credit)

		// Get pet account status and usage
		pet.GET("/status/:pet_id", petHandler.GetStatus)

		// Suspend a pet's compute access
		pet.POST("/suspend/:pet_id", petHandler.Suspend)

		// Reactivate a suspended pet
		pet.POST("/reactivate/:pet_id", petHandler.Reactivate)

		// Batch query usage for multiple pets
		pet.POST("/usage/batch", petHandler.BatchUsage)
	}
}
