package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func runIftop(device_name string) string {
	// Run iftop command and get output
	cmd := exec.Command("iftop", "-i", device_name, "-B", "-t", "-P", "-n", "-N", "-s", "2")
	output, err := cmd.CombinedOutput()
	output_str := string(output)
	fmt.Println(output_str)
	if err != nil {
		fmt.Println("Run iftop error: ", err)
		return ""
	}
	return output_str
}

func getDB() (*sql.DB, error) {
	now := time.Now().Format("2006-01-02")
	_, err := os.Stat(now)
	if os.IsNotExist(err) {
		fmt.Println("Create folder: ", now)
		err := os.Mkdir(now, 0755)
		if err != nil {
			fmt.Println("Create folder failed" + err.Error())
			return nil, err
		}
	}
	//
	db_file_path := now + "/network_detail.db"
	db, err := sql.Open("sqlite3", db_file_path)
	if err != nil {
		fmt.Println("Open database error：", err)
		return nil, err
	}
	//
	if !checkTableExists(db) {
		err = createTable(db)
		if err != nil {
			return nil, err
		}
	}
	// defer db.Close()
	return db, nil
}

func checkTableExists(db *sql.DB) bool {
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name='network_detail'"
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Query table name error:", err)
		return false
	}
	defer rows.Close()
	//
	return rows.Next()
}

func createTable(db *sql.DB) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS network_detail (
			id INTEGER PRIMARY KEY,
			src_host VARCHAR(128),
			src_port INTEGER,
			tgt_host VARCHAR(128),
			tgt_port INTEGER,
			send_bytes INTEGER,
			receive_bytes INTEGER,
			time VARCHAR(32)
		);
	`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		fmt.Println("Table create error：", err)
		return err
	}
	fmt.Println("Table create success")
	return nil
}

func insertData(src_host string, src_port int, tgt_host string, tgt_port int, send_bytes float64, receive_bytes float64, time string) {
	db, err := getDB()
	if err != nil {
		return
	}
	defer db.Close()
	// Insert data to table
	insertDataSQL := `
		INSERT INTO network_detail (src_host, src_port, tgt_host, tgt_port, send_bytes, receive_bytes, time)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	_, err = db.Exec(insertDataSQL, src_host, src_port, tgt_host, tgt_port, send_bytes, receive_bytes, time)
	if err != nil {
		fmt.Println("Insert data error：", err)
		return
	}
	fmt.Println("Insert data success")
}

func parseBytesString(str string) float64 {
	// Use regex to parse num and unit from string
	pattern := `([\d\.]+)([GMKB]+)`
	regex := regexp.MustCompile(pattern)
	result := regex.FindAllStringSubmatch(str, -1)
	// Regex query to variables
	slice := result[0][1:]
	num, unit := slice[0], slice[1]
	value, _ := strconv.ParseFloat(num, 64)
	// Unit to bytes
	mapping := map[string]int{
		"B":  1,
		"KB": 1000,
		"MB": 1000000,
		"GB": 1000000,
	}
	value = value * float64(mapping[unit])
	//
	return value
}

func parseCommandOut(text string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	// Use regex to parse output text of command iftop
	pattern := `\s+([\d\.]+)\:(\d+)\s+=>\s+([\d\.GMKB]+)\s+([\d\.GMKB]+)\s+([\d\.GMKB]+)\s+([\d\.GMKB]+)\n\s+([\d\.]+):(\d+)\s+<=\s+([\d\.GMKB]+)\s+([\d\.GMKB]+)\s+([\d\.GMKB]+)\s+([\d\.GMKB]+)`
	regex := regexp.MustCompile(pattern)
	result := regex.FindAllStringSubmatch(text, -1)
	for _, match := range result {
		slice := match[1:]
		fmt.Println(slice)
		// use column last 2s to calculate speed
		send_bytes := parseBytesString(slice[2]) / 2
		receive_bytes := parseBytesString(slice[8]) / 2
		// port column to int
		src_port, _ := strconv.Atoi(slice[1])
		tgt_port, _ := strconv.Atoi(slice[7])
		// save data
		insertData(slice[0], src_port, slice[6], tgt_port, send_bytes, receive_bytes, now)
	}
}

func main() {
	// Parse command line args
	var device string
	flag.StringVar(&device, "i", "enp3s0", "device interface name")
	flag.Parse()
	// Run loop
	for {
		text := runIftop(device)
		parseCommandOut(text)
	}
}
