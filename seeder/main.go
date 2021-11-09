package seeder

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func Seed() {
	os.Remove("./sqlite.db")

	db, err := sql.Open("sqlite3", "./sqlite.db")
	checkErr(err)

	createUsersTable(db)
	createUsers(db)
	createHighscoreTable(db)
	createHighscores(db)

	db.Close()
}

func createUsersTable(db *sql.DB) {
	stmt, err := db.Prepare("CREATE TABLE `users` (`id` INTEGER PRIMARY KEY AUTOINCREMENT, `email` VARCHAR(64), `password` VARCHAR(64));")
	checkErr(err)

	_, err = stmt.Exec()
	checkErr(err)
}

func createUsers(db *sql.DB) {
	stmt, err := db.Prepare("INSERT INTO users(id, email, password) values(?, ?, ?)")
	checkErr(err)

	_, err = stmt.Exec(1, "root@vulnerable-click.game", "adminPassword")
	checkErr(err)

	_, err = stmt.Exec(2, "user@vulnerable-click.game", "userPassword")
	checkErr(err)
}

func createHighscoreTable(db *sql.DB) {
	stmt, err := db.Prepare("CREATE TABLE `highscore` (`id` INTEGER PRIMARY KEY AUTOINCREMENT, `userid` INTEGER, `points` INTEGER, `victoryshout` VARCHAR(256) NULL);")
	checkErr(err)

	_, err = stmt.Exec()
	checkErr(err)
}

func createHighscores(db *sql.DB) {
	stmt, err := db.Prepare("INSERT INTO highscore(id, userid, points, victoryshout) values(?, ?, ?, ?)")
	checkErr(err)

	_, err = stmt.Exec(1, 1, 42, "The answer to the Ultimate Question of Life, the Universe, and Everything")
	checkErr(err)

	_, err = stmt.Exec(2, 1, 1337, "Leet!")
	checkErr(err)

	_, err = stmt.Exec(3, 2, 7, "Not so bad. Is it?")
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
