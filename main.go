package main

import (
	"github.com/notblessy/go-ingnerd/src/models"
	"github.com/notblessy/go-ingnerd/src/routes"
	"gorm.io/gorm"
)

var (
	db *gorm.DB = models.ConnectDB()
)

func main() {
	defer models.DisconnectDB(db)

	routes.Routes()
}
