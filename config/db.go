package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	   "gorm.io/gorm/schema"
)
var DB *gorm.DB
var err2 error
func Connect(){

err:=godotenv.Load(".env")
if err!=nil{
	log.Fatalf("error getting data from .env:%v",err)

}
dsn:=fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",os.Getenv("HOST"),os.Getenv("PORT"),os.Getenv("user_name"),os.Getenv("password"),os.Getenv("db_name"),os.Getenv("ssl"))
DB,err2=gorm.Open(postgres.Open(dsn),&gorm.Config{SkipDefaultTransaction:true ,Logger: logger.Default.LogMode(logger.Warn),DisableForeignKeyConstraintWhenMigrating:true,    NamingStrategy: schema.NamingStrategy{
        SingularTable: true,
		NoLowerCase: true,
    },
})
// DB=Db
// DB
DB = DB.Debug()

if err2!=nil{
	log.Fatalf("error connnecting database:%v",err2)
}
fmt.Println("connected to database")
}