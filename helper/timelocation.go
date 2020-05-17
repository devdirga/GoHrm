package helper

import (
	"time"
)

func TimeLocation(timezone string) (string, error) {
	t := time.Now()

	lc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", err
	}

	now := t.In(lc)

	return now.Format("2006-01-02 15:04"), nil

}
func TimeLocationUnFormat(timezone string) (time.Time, error) {
	t := time.Now()

	lc, err := time.LoadLocation(timezone)
	if err != nil {
		return t, err
	}

	now := t.In(lc)

	return now, nil

}

func ExpiredDateTime(timezone string, isOvertime bool) (string, error) {
	t := time.Now().Add(6 * time.Hour)
	if isOvertime {
		t = time.Now().Add(12 * time.Hour)
	}

	// fmt.Println("------------- t ", t)

	lc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", err
	}

	now := t.In(lc)

	// fmt.Println("------------- t ", now)

	return now.Format("02-01-2006 15:04"), nil
}

func ExpiredRemaining(timezone string, isOvertime bool) (string, error) {
	t := time.Now().Add(5*time.Hour + 15*time.Minute)
	if isOvertime {
		t = time.Now().Add(11*time.Hour + 15*time.Minute)
	}

	// fmt.Println("------------- t ", t)

	lc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", err
	}

	rem := t.In(lc)

	// fmt.Println("------------- t ", now)

	return rem.Format("02-01-2006 15:04"), nil
}

func IsMoreAYear(jointdate string) (ayear bool, year int, month int, day int) {
	now := time.Now()
	start, _ := time.Parse("2006-01-02", jointdate)
	// fmt.Println("start year ------- ", start)
	y, m, d, _, _, _ := GetComparisonDate(start, now)

	if y > 1 {
		return true, y, m, d
	}

	return false, y, m, d
}

func GetComparisonDate(start, now time.Time) (year, month, day, hour, min, sec int) {
	if start.Location() != now.Location() {
		now = now.In(start.Location())
	}
	if start.After(now) {
		start, now = now, start
	}
	y1, M1, d1 := start.Date()

	y2, M2, d2 := now.Date()
	// fmt.Println("now year2 ------- ", int(start.Year()))
	// fmt.Println("now year2 ------- ", y2)

	h1, m1, s1 := start.Clock()
	h2, m2, s2 := now.Clock()

	year = int(y2 - start.Year())
	// fmt.Println("--------------- year ", year)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}

	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	// fmt.Println("--------------- year2 ", year)

	return
}

func TimeLocationLog(timezone string) (string, error) {
	t := time.Now()

	lc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", err
	}

	now := t.In(lc)

	return now.Format("02-01-2006 15:04:01"), nil

}
