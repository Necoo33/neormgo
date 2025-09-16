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

func TestTransaction(t *testing.T) {
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

	fmt.Println("Transaction status on connection moment: ", db.Tx != nil)

	err = db.Begin()
	if err != nil {
		t.Fatalf("Begin başarısız: %s", err)
	}

	fmt.Println("Transaction status on first step: ", db.Tx != nil)

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

	fmt.Printf("Here is your result of first step, random column: %v\n", rows[0][randomColumn])

	length := db.Count(table)
	//length.Where("id", ">", 10)

	fmt.Println("Transaction status on second step: ", db.Tx != nil)

	err = length.Execute()

	if err != nil {
		t.Fatalf("Error occured when we try to fetch data: %s", err)
	}

	lengthOfUsers := length.Length()

	fmt.Printf("Here is your result of second step, rows length: %d\n", lengthOfUsers)

	database = db.Select([]string{"age"})
	database.Table(table)
	database.Finish()

	fmt.Println("Transaction status on third step: ", db.Tx != nil)

	err = database.Execute()
	if err != nil {
		t.Fatalf("Error occured when we try to fetch data: %s", err)
	}

	rows, err = database.Rows()
	if err != nil {
		t.Fatalf("Error occured when we try to get rows: %s", err)
	}

	fmt.Printf("Here is your result of transaction, age column: %v\n", rows[1]["age"])

	err = db.Commit()
	if err != nil {
		t.Fatalf("Commit başarısız: %s", err)
	}

	fmt.Println("Transaction status on end: ", db.Tx != nil)

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
