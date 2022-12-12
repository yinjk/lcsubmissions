// @Desc
// @Author  inori
// @Update
package models

import "gorm.io/gorm"

type AuditLog struct {
	gorm.Model
	Ip        string `json:"ip"`
	Timestamp int64  `json:"timestamp"`
	Operator  string `json:"operator"`
	Param     string `json:"param"`
	Path      string `json:"path"`
	TimeCost  string `json:"timeCost"`
}

func (a *AuditLog) TableName() string {
	return "audit_log"
}

type Submission struct {
	gorm.Model
	QuestionId      string `json:"questionId"`
	Title           string `json:"title"`
	TranslatedTitle string `json:"translatedTitle"`
	TitleSlug       string `json:"titleSlug"`
	Status          string `json:"status"`
	Lang            string `json:"lang"`
	LCId            string `gorm:"unique_index:user_one_submit" json:"lcId"`
	SubmitTime      int64  `gorm:"unique_index:user_one_submit" json:"submitTime"`
}

func (s *Submission) TableName() string {
	return "submission"
}
