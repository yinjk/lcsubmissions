// @Desc
// @Author  inori
// @Update
package handler

import (
	"errors"
	"gorm.io/gorm"
	"lcsubmissions/dao"
	"lcsubmissions/models"
	"log"
)

func QuerySubmissions(from, to int64, lcId string) []models.Submission {
	var res = make([]models.Submission, 0)
	query := "submit_time >= ? and submit_time <= ?"
	args := []interface{}{from, to}
	if lcId != "" {
		query += " and lc_id = ?"
		args = append(args, lcId)
	}
	if err := dao.DB.Where(query, args...).Find(&res).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("query submission with args %++v failed, err = %v\n", args, err)
		return res
	}
	return res
}
