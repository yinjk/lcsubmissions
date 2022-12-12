// @Desc
// @Author  inori
// @Update
package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"lcsubmissions/dao"
	"lcsubmissions/models"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var userIdMap = make(map[string]string)
var usernameMap = make(map[string]string)

func Init(users [][]string, interval time.Duration) {
	for _, user := range users {
		if len(user) == 2 {
			usernameMap[user[0]] = user[1]
			userIdMap[user[1]] = user[0]
		}
	}
	StartRefreshSubmissions(users, interval)
}
func ListenAndStart(port int, users [][]string, interval time.Duration) {
	Init(users, interval)
	r := gin.Default()
	r.GET("/leetcode", func(context *gin.Context) {
		Leetcode(context, "day")
	})
	r.GET("/leetcode/week", func(context *gin.Context) {
		Leetcode(context, "week")
	})
	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		panic(err)
	}
}
func Leetcode(c *gin.Context, typ string) {
	now := time.Now()
	queryDay := c.Query("day")
	requestParam := make(url.Values)
	defer func() {
		ip := RemoteIp(c.Request)
		timeCost := time.Now().Sub(now).String()
		log.Printf("ip:[%s] requested with param: %++v cost time: %s\n", ip, requestParam, timeCost)
		audit := models.AuditLog{
			Ip:        ip,
			Timestamp: now.Unix(),
			Operator:  "",
			Param:     fmt.Sprintf("%v", requestParam),
			Path:      c.FullPath(),
			TimeCost:  timeCost,
		}
		dao.DB.Create(&audit)
	}()

	switch typ {
	case "day":
		c.String(http.StatusOK, StatisticDay(queryDay))
	case "week":
		c.String(http.StatusOK, StatisticWeek(queryDay))
	}
}
func StatisticDay(queryDay string) string {
	now := time.Now().Truncate(time.Hour)
	day := now.Add(-time.Duration(now.Hour()) * time.Hour)
	if queryDay != "" {
		parse, err := time.Parse("20060102", queryDay)
		if err == nil {
			day = parse
		}
	}
	title := fmt.Sprintf("%s (day)", day.Format("2006-01-02"))
	return StatisticText(day.Unix(), day.Add(24*time.Hour).Unix(), title)
}
func StatisticWeek(queryDay string) string {
	now := time.Now().Truncate(time.Hour)
	day := now.Add(-time.Duration(now.Hour()) * time.Hour)
	if queryDay != "" {
		parse, err := time.Parse("20060102", queryDay)
		if err == nil {
			day = parse
		}
	}
	weekDay := int(day.Weekday())
	if weekDay == 0 {
		weekDay = 7
	}
	weekDay--
	monday := day.AddDate(0, 0, -weekDay)
	title := fmt.Sprintf("%s(week)", monday.Format("2006-01-02"))
	return StatisticText(monday.Unix(), monday.AddDate(0, 0, 7).Unix(), title)
}

func StatisticText(from, to int64, formatStr string) string {
	sb := strings.Builder{}
	submissions := QuerySubmissions(from, to, "")
	var submissionMap = make(map[string][]models.Submission)
	for k, _ := range userIdMap {
		submissionMap[k] = make([]models.Submission, 0)
	}
	for _, submission := range submissions {
		s := submissionMap[submission.LCId]
		s = append(s, submission)
		submissionMap[submission.LCId] = s
	}

	var statistics = make([]Statistics, 0)
	for lcId, submissionsI := range submissionMap {
		s := GetStatistics(submissionsI)
		s.Id = lcId
		s.UserName = userIdMap[lcId]
		statistics = append(statistics, s)
	}

	sb.WriteString(fmt.Sprintf("%*s\n", 44, fmt.Sprintf("lastest sync time: %s", GetLatestSyncTime())))
	sb.WriteString(fmt.Sprintf("============= %s =============\n", formatStr))
	sb.WriteString(fmt.Sprintf("%10s | %5s | %5s | %3s | #0 \n", "id", "提交数", "通过率", "刷题数"))
	sort.SliceStable(statistics, func(i, j int) bool {
		if statistics[i].QuestionCount == statistics[j].QuestionCount {
			if statistics[i].SubmitCount == statistics[j].SubmitCount {
				if statistics[i].ThroughPercent == statistics[j].ThroughPercent {
					return statistics[i].UserName > statistics[j].UserName
				}
				return statistics[i].ThroughPercent > statistics[j].ThroughPercent
			}
			return statistics[i].SubmitCount > statistics[j].SubmitCount
		}
		return statistics[i].QuestionCount > statistics[j].QuestionCount
	})
	for i, s := range statistics {
		if i == 0 {
			sb.WriteString(fmt.Sprintf("%10s | %7d | %6.2f%% | %5d | #%-2d🥇\n", s.UserName, s.SubmitCount, s.ThroughPercent, s.QuestionCount, i+1))
		} else {
			sb.WriteString(fmt.Sprintf("%10s | %7d | %6.2f%% | %5d | #%-2d\n", s.UserName, s.SubmitCount, s.ThroughPercent, s.QuestionCount, i+1))
		}
	}
	return sb.String()
}
func GetStatistics(submissions []models.Submission) Statistics {
	//status:
	// A_10 通过
	// A_11 解答错误
	// A_15 执行出错
	// A_14 超出执行时间
	var s Statistics
	var sucMap = make(map[string]bool)
	success := 0
	for _, submission := range submissions {
		if submission.Status == "A_10" {
			success++
			sucMap[submission.QuestionId] = true
		}
	}
	throughPercent := float64(0)
	if len(submissions) > 0 {
		throughPercent = float64(success) * 100 / float64(len(submissions))
	}
	s.SubmitCount = len(submissions)
	s.ThroughPercent = throughPercent
	s.QuestionCount = len(sucMap)
	return s
}

type Statistics struct {
	UserName       string
	Id             string
	SubmitCount    int
	ThroughPercent float64
	QuestionCount  int
}

func RemoteIp(req *http.Request) string {
	var remoteAddr string
	// RemoteAddr
	remoteAddr = req.RemoteAddr
	if remoteAddr != "" {
		return remoteAddr
	}
	// ipv4
	remoteAddr = req.Header.Get("ipv4")
	if remoteAddr != "" {
		return remoteAddr
	}
	//
	remoteAddr = req.Header.Get("XForwardedFor")
	if remoteAddr != "" {
		return remoteAddr
	}
	// X-Forwarded-For
	remoteAddr = req.Header.Get("X-Forwarded-For")
	if remoteAddr != "" {
		return remoteAddr
	}
	// X-Real-Ip
	remoteAddr = req.Header.Get("X-Real-Ip")
	if remoteAddr != "" {
		return remoteAddr
	} else {
		remoteAddr = "127.0.0.1"
	}
	return remoteAddr
}
