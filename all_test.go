package neormgo

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// change these variables to test your own database:

var database = "postgres"
var table = "users"
var randomColumn = "name"
var resultAlias = "result"
var callType = "function"

func TestConnection(t *testing.T) {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatalf("`.env` dosyası yüklenemedi: %s", err)
	}

	db := Neorm{}
	dbConnURL := os.Getenv("DB_CONN_URL")

	db, err = db.Connect(dbConnURL, database)
	if err != nil {
		t.Fatalf("Connect başarısız: %s", err)
	}

	if err = db.Pool.Ping(); err != nil {
		t.Fatalf("Ping başarısız: %v", err)
	}
}

func TestSelect(t *testing.T) {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatalf("`.env` dosyası yüklenemedi: %s", err)
	}

	db := Neorm{}
	dbConnURL := os.Getenv("DB_CONN_URL")

	db, err = db.Connect(dbConnURL, database)
	if err != nil {
		t.Fatalf("Connect başarısız: %s", err)
	}

	database := db.Select("*")
	database.Table(table)
	database.Finish()

	err = database.Execute()

	if err != nil {
		t.Fatalf("Error occured when we try to fetch data: %s", err)
	}

	rows, err := database.Rows()

	if err != nil {
		t.Fatalf("Error occured when we try to get rows: %s", err)
	}

	fmt.Printf("Here is your rows count: %d\n", len(rows))

	fmt.Printf("Here is a random data: %v\n", rows[0][randomColumn])
}

func TestLength(t *testing.T) {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatalf("`.env` dosyası yüklenemedi: %s", err)
	}

	db := Neorm{}
	dbConnURL := os.Getenv("DB_CONN_URL")

	db, err = db.Connect(dbConnURL, database)
	if err != nil {
		t.Fatalf("Connect başarısız: %s", err)
	}

	length := db.Count(table)
	//length.Where("id", ">", 10)
	err = length.Execute()

	if err != nil {
		t.Fatalf("Error occured when we try to fetch data: %s", err)
	}

	lengthOfUsers := length.Length()

	fmt.Printf("Here is your rows length: %d\n", lengthOfUsers)
}

func TestProcedureCall(t *testing.T) {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatalf("`.env` dosyası yüklenemedi: %s", err)
	}

	db := Neorm{}
	dbConnURL := os.Getenv("DB_CONN_URL")

	db, err = db.Connect(dbConnURL, database)
	if err != nil {
		t.Fatalf("Connect başarısız: %s", err)
	}

	procedure := db.Call(callType, "get_mock_result", resultAlias)
	procedure.Finish()

	err = procedure.Execute()
	if err != nil {
		t.Fatalf("Error occured when we try to fetch data: %s", err)
	}

	rows, err := procedure.Rows()

	if err != nil {
		t.Fatalf("Error occured when we try to get rows: %s", err)
	}

	fmt.Printf("Here is your result of procedure call: %v\n", rows[0][resultAlias])

}
