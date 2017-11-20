package main

import (
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type User struct {
 ID uint
 Username string
 Password string
}

var db *gorm.DB

func dbInit() {
 var err error
 db, err = gorm.Open("sqlite3", "./gorm.db")
 if err != nil {
  panic("failed to connect database")
 }
 //defer db.Close()

 db.AutoMigrate(&User{})
 db.AutoMigrate(&Media{})
}
