package main

import (
	"aplikasi-bioskop/controllers"
	"aplikasi-bioskop/routers"
)

func main() {
	controllers.ConnectDB()
	var PORT = ":8080"

	routers.StartServer().Run(PORT)
}
