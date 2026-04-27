package models

import (
	"gorm.io/driver/sqlite"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {

        database, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})

        if err != nil {
                panic("Failed to connect to database!")
        }

        err = database.AutoMigrate(&User{})
        if err != nil {
                return
        }
        err = database.AutoMigrate(&Project{})
        if err != nil {
                return
        }
        err = database.AutoMigrate(&Instance{})
        if err != nil {
                return
        }
        


        DB = database
}