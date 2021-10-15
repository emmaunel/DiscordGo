package util

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/emmaunel/DiscordGo/pkg/agent"
	"github.com/emmaunel/DiscordGo/pkg/agents"

	"github.com/emmaunel/DiscordGo/pkg/util/constants"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB
var mutex = &sync.Mutex{}
var err error

func CreateDatabaseAndTable() {
	DB, err = sql.Open("mysql", constants.DBUsername+":"+constants.DBPassword+"@(localhost:3306)/")

	// if there is an error opening the connection, handle it
	if err != nil {
		panic(err.Error())
	}

	// defer DB.Close()

	// Create the database if it doesnt'exist
	_, err = DB.Exec("CREATE DATABASE Agents")
	if err != nil {
		println("Database already exist 👍")
	}

	_, err = DB.Exec("USE Agents")
	if err != nil {
		fmt.Println(err.Error())
	}

	// Create table
	tableCreation, err := DB.Prepare("CREATE TABLE Agent(UUID VARCHAR(45) NOT NULL,Hostname VARCHAR(45) NOT NULL,OS VARCHAR(45) NOT NULL,IP VARCHAR(45) NOT NULL, Status VARCHAR(45) NOT NULL DEFAULT 'Alive',TimeStamp VARCHAR(45), PRIMARY KEY (UUID));")
	// tableCreation, err := DB.Prepare("CREATE TABLE Agent(UUID VARCHAR(45) NOT NULL,Hostname VARCHAR(45) NOT NULL,OS VARCHAR(45) NOT NULL,IP VARCHAR(45) NOT NULL, Status VARCHAR(45) NOT NULL DEFAULT 'Alive', Timestamp DATETIME NULL,PRIMARY KEY (UUID));")

	if err != nil {
		fmt.Println(err.Error())
	}

	_, err = tableCreation.Exec()
	if err != nil {
		println("Table already exist... 👍")
	} else {
		fmt.Println("Table created successfully... 👍")
	}
}

func InsertAgentToDB(agentID string, hostname string, agentIP string, agentOS string) {

	// if DB == nil {
	mutex.Lock()
	defer mutex.Unlock()
	if DB == nil {
		fmt.Println("Creating Single Instance Now")
		// var err error
		//insert statement
		insertStat, err := DB.Prepare("INSERT INTO Agent VALUES (?, ?, ?, ?, ?, ?)")
		if err != nil {
			panic(err.Error())
		}
		// execute
		_, _ = insertStat.Exec(agentID, hostname, agentOS, agentIP, "Alive", "null")
		if err != nil {
			panic(err)
		}
	}
	// }

	// //insert statement
	// insertStat, err := DB.Prepare("INSERT INTO Agent VALUES (?, ?, ?, ?, ?, ?)")
	// if err != nil {
	// 	panic(err.Error())
	// }
	// // execute
	// _, _ = insertStat.Exec(agentID, hostname, agentOS, agentIP, "Alive", "null")
}

func LoadFromDB() {
	result, err := DB.Query("SELECT * FROM Agents.agent")
	if err != nil {
		panic(err) // proper error handling instead of panic in your app
	}
	var id, hostname, os, ip, status, timestamp string
	for result.Next() {
		_ = result.Scan(&id, &hostname, &os, &ip, &status, &timestamp)
		agents.Agents = append(agents.Agents, &agent.Agent{UUID: id, HostName: hostname, OS: os, IP: ip, Status: status, Timestamp: timestamp})
	}
}

func UpdateAgentTimestamp(agentID string, timestamp string) {
	updateStat, err := DB.Prepare("update Agents.agent set TimeStamp=? where UUID=?")
	if err != nil {
		panic(err) // proper error handling instead of panic in your app
	}
	_, err = updateStat.Exec(timestamp, agentID)
	if err != nil {
		fmt.Println(err)
	}

}

func UpdateAgentStatus(agentID string, status string) {
	updateStat, err := DB.Prepare("update Agents.agent set Status=? where UUID=?")
	if err != nil {
		panic(err) // proper error handling instead of panic in your app
	}
	_, err = updateStat.Exec(status, agentID)
	if err != nil {
		fmt.Println(err)
	}
}

func DropDB() {
	updateStat, err := DB.Prepare("drop database Agents")
	if err != nil {
		panic(err) // proper error handling instead of panic in your app
	}
	_, err = updateStat.Exec()
	if err != nil {
		fmt.Println(err)
	}
}

func RemoveAgentFromDB(agentID string) {
	// delete data
	stmt, _ := DB.Prepare("delete from Agents.agent where UUID=?")
	_, _ = stmt.Exec(agentID)
}

func AgentsByHostname(hostname string) []agent.Agent {
	return nil
}
