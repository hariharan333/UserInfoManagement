package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/notblessy/go-ingnerd/src/controllers"
)

func Routes() {
	route := gin.Default()

	route.POST("/user/create", controllers.CreateUser)
	route.GET("/user/getall", controllers.GetAllUsers)
	route.GET("/user/get/:userid", controllers.GetUser)
	route.PUT("/user/update/:userid", controllers.UpdateUser)
	route.DELETE("/user/delete/:userid", controllers.DeleteUser)

	route.Run()
}
