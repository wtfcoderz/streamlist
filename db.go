package main

import (
//    "fmt"
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

 // User
 db.AutoMigrate(&User{})
 //var count int
 //db.Model(&User{}).Count(&count)
 //fmt.Println(count)
 //u1 := User{Username: "admin", Password: "admin"}
 //u2 := User{Username: "ro", Password: "ro"}
 //db.Create(&u1)
 //var u3 User // identify a Person type for us to store the results in
 //db.First(&u3) // Find the first record in the Database and store it in p3
 //fmt.Println(u1.Username)
 //fmt.Println(u2.Username)
 //fmt.Println(u3.Username) // print out our record from the database
}
