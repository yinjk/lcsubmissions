// @Desc
// @Author  inori
// @Update
package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"lcsubmissions/dao"
	"lcsubmissions/models"
	"log"
	"net/http"
	"time"
)

var (
	lcGraphUrl        = "https://leetcode-cn.com/graphql"
	submissionReqJson = `{
    "operationName": "recentSubmissions",
    "variables": {"userSlug":"%s"},
    "query": "query recentSubmissions($userSlug: String!){recentSubmissions(userSlug: $userSlug){status lang question{questionFrontendId title difficulty translatedTitle titleSlug __typename}submitTime __typename}}"
}`
)
var syncTime time.Time

func GetLatestSyncTime() string {
	return syncTime.Format("15:04:05")
}

func StartRefreshSubmissions(users [][]string, interval time.Duration) {
	//RefreshSubmissions(users)
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("start refresh users submission")
				RefreshSubmissions(users)
				log.Println("refresh users submission finished")
			}
		}
	}()
}

func RefreshSubmissions(users [][]string) {
	for _, val := range users {
		if len(val) > 0 {
			submissions, err := GetUserSubmissions(val[1])
			if err != nil {
				log.Printf("get user[%s] submissions failed, err = %v", val[1], err)
				continue
			}
			for i, _ := range submissions {
				if err := dao.DB.FirstOrCreate(&submissions[i], "submit_time = ? and lc_id = ?", submissions[i].SubmitTime, submissions[i].LCId).Error; err != nil {
					log.Printf("first or create \n")
					continue
				}
			}
		}
	}
	syncTime = time.Now()
}

func GetUserSubmissions(lcId string) ([]models.Submission, error) {
	var submissions = make([]models.Submission, 0)
	reqBody := []byte(fmt.Sprintf(submissionReqJson, lcId))
	req, _ := http.NewRequest("GET", lcGraphUrl, bytes.NewBuffer(reqBody))
	//    head = {'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/5'
	//                          ''}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
	req.Header.Set("Content-Type", "application/json")
	get, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("get userinfo failed, err = %v \n", err)
		return submissions, err
	}
	body, err := ioutil.ReadAll(get.Body)
	if err != nil {
		log.Printf("read user body failed, err = %v \n", err)
		return submissions, err
	}
	var res SubmissionRes
	if err := json.Unmarshal(body, &res); err != nil {
		log.Printf("unmarshal body failed, err = %v \n", err)
		return submissions, err
	}
	for _, submission := range res.Data.RecentSubmissions {
		submissions = append(submissions, models.Submission{
			LCId:            lcId,
			QuestionId:      submission.Question.QuestionFrontendId,
			Title:           submission.Question.Title,
			Difficulty:      submission.Question.Difficulty,
			TranslatedTitle: submission.Question.TranslatedTitle,
			TitleSlug:       submission.Question.TitleSlug,
			Status:          submission.Status,
			Lang:            submission.Lang,
			SubmitTime:      submission.SubmitTime,
		})
	}
	return submissions, nil
}

type SubmissionRes struct {
	Data struct {
		RecentSubmissions []struct {
			Status   string `json:"status"`
			Lang     string `json:"lang"`
			Question struct {
				QuestionFrontendId string `json:"questionFrontendId"`
				Title              string `json:"title"`
				Difficulty         string `json:"difficulty"`
				TranslatedTitle    string `json:"translatedTitle"`
				TitleSlug          string `json:"titleSlug"`
				Typename           string `json:"__typename"`
			} `json:"question"`
			SubmitTime int64  `json:"submitTime"`
			Typename   string `json:"__typename"`
		} `json:"recentSubmissions"`
	} `json:"data"`
}
type SimpleSubmission struct {
	QuestionId      string `json:"questionId"`
	TranslatedTitle string `json:"translatedTitle"`
	Status          string `json:"status"`
	Lang            string `json:"lang"`
}
