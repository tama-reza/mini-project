package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Mendaftarkan driver postgres
)

type Bioskop struct {
	ID     int     `json:"id"`
	Nama   string  `json:"nama"`
	Lokasi string  `json:"lokasi"`
	Rating float64 `json:"rating"`
}

var db *sql.DB

// Pintu Komunikasi ke postgresql
func ConnectDB() {
	// Ambil dan muat file .env ke dalam sistem aplikasi Go
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Peringatan: File .env tidak ditemukan, menggunakan nilai default.")
	}

	psqlInfo := os.Getenv("DATABASE_URL")

	if psqlInfo == "" {
		// Ambil data konfigurasi dari file .env
		host := getEnv("PGHOST", "postgres.railway.internal")
		user := getEnv("PGUSER", "postgres")
		password := getEnv("PGPASSWORD", "")
		dbname := getEnv("PGDATABASE", "railway")

		// Konversi teks port di .env menjadi angka integer untuk fmt.Sprintf
		portStr := getEnv("PGPORT", "5432")
		port, _ := strconv.Atoi(portStr)

		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	}

	// 1. Open Database Connection
	db, err = sql.Open("postgres", psqlInfo) // mengembalikan 2 return, db dan err
	if err != nil {
		panic(fmt.Sprintf("Error opening database: %v", err))
	}

	// 2. Ping database to verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		panic(fmt.Sprintf("Database unreachable: %v", err))
	}

	fmt.Println("Succesfully connected to database")

	// Membuat tabel otomatis jika belum ada di database Railway
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS bioskop (
		id SERIAL PRIMARY KEY,
		nama VARCHAR(255) NOT NULL,
		lokasi TEXT NOT NULL,
		rating NUMERIC(3,2) NOT NULL
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		panic(fmt.Sprintf("Gagal membuat tabel otomatis: %v", err))
	}
	fmt.Println("Tabel 'bioskop' siap digunakan atau sudah ada.")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func CreateBioskop(ctx *gin.Context) {
	var newBioskop Bioskop

	if err := ctx.ShouldBindJSON(&newBioskop); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ID otomatis di-generate oleh SERIAL di Postgres dan dikembalikan lewat RETURNING id
	sqlStatement := `INSERT INTO bioskop (nama, lokasi, rating) 
	VALUES ($1, $2, $3) 
	RETURNING id`

	err := db.QueryRow(sqlStatement, newBioskop.Nama, newBioskop.Lokasi, newBioskop.Rating).Scan(&newBioskop.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan data ke database"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"bioskop": newBioskop,
	})
}

func UpdateBioskop(ctx *gin.Context) {
	id := ctx.Param("id") // Mengambil ID dari URL, misal: /bioskop/1
	var updateData Bioskop

	idInt, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID harus berupa angka yang valid"})
		return
	}

	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateData.ID = idInt

	sqlStatement := `UPDATE bioskop SET nama = $1, lokasi = $2, rating = $3 WHERE id = $4`

	res, err := db.Exec(sqlStatement, updateData.Nama, updateData.Lokasi, updateData.Rating, idInt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui data"})
		return
	}

	// Memeriksa apakah ada baris yang berubah (apakah ID tersebut ada di DB)
	count, _ := res.RowsAffected()
	if count == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Data bioskop tidak ditemukan"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Data berhasil diperbarui",
	})
}

func GetBioskop(ctx *gin.Context) {
	// Menampung data secara dinamis dari database (Bukan variabel global)
	var listBioskop []Bioskop

	sqlStatement := `SELECT id, nama, lokasi, rating FROM bioskop ORDER BY id ASC`

	rows, err := db.Query(sqlStatement)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var b Bioskop
		err := rows.Scan(&b.ID, &b.Nama, &b.Lokasi, &b.Rating)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca baris data"})
			return
		}
		listBioskop = append(listBioskop, b)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": listBioskop,
	})
}

func DeleteBioskop(ctx *gin.Context) {
	id := ctx.Param("id")

	// Tambahkan validasi angka di sini
	idInt, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID harus berupa angka yang valid"})
		return
	}

	sqlStatement := `DELETE FROM bioskop WHERE id = $1`

	res, err := db.Exec(sqlStatement, idInt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus data"})
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Data bioskop tidak ditemukan"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Bioskop dengan ID %d berhasil dihapus", idInt),
	})
}
