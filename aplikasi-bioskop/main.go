package main

import (
	"aplikasi-bioskop/controllers"
	"aplikasi-bioskop/routers"
	"os"
)

func main() {
	controllers.ConnectDB()

	// 1. Ambil port dari sistem Railway
	port := os.Getenv("PORT")

	// 2. Jika dijalankan di lokal, os.Getenv("PORT") akan kosong.
	// Kita beri nilai cadangan (fallback) ke "8080"
	if port == "" {
		port = "8080"
	}

	// 3. Jalankan server dengan format ":port" (misal: :8080 atau :3421)
	routers.StartServer().Run(":" + port)
}
