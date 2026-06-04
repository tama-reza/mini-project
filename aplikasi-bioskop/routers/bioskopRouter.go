package routers

import (
	"aplikasi-bioskop/controllers"

	"github.com/gin-gonic/gin"
)

func StartServer() *gin.Engine {
	router := gin.Default()

	router.GET("/bioskop", controllers.GetBioskop)
	router.POST("/bioskop", controllers.CreateBioskop)
	router.PUT("/bioskop/:id", controllers.UpdateBioskop)
	router.DELETE("/bioskop/:id", controllers.DeleteBioskop)

	return router
}
