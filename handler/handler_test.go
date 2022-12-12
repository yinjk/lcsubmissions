// @Desc
// @Author  inori
// @Update
package handler

import (
	"fmt"
	"testing"
	"time"
)

func TestWeek(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	now = now.Add(time.Duration(-now.Hour()) * time.Hour)
	fmt.Println(now)
	weekDay := int(now.Weekday())
	if weekDay == 0 {
		weekDay = 7
	}
	weekDay--
	monday := now.AddDate(0, 0, -weekDay)
	fmt.Println(monday)
	fmt.Println(monday.AddDate(0, 0, 7))
}
