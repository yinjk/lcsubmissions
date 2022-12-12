// @Desc
// @Author  inori
// @Update
package dao

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"lcsubmissions/models"
)

type DBConfig struct {
	Addr     string
	Database string
	Username string
	Password string
}

var DB *gorm.DB

func InitMysql(dsn string) (err error) {
	//dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := DB.AutoMigrate(&models.AuditLog{}, &models.Submission{}); err != nil {
		return err
	}
	return nil
}
