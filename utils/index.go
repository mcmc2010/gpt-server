package utils

import (
	"math"
	"math/rand"
	"regexp"
	"time"
)

// TIME
const (
	TIME_SECOND   = 1.0
	TIME_MINUTE   = 1.0 * 60.0
	TIME_HOUR     = 1.0 * 60.0 * 60.0
	TIME_DAY      = 1.0 * 60.0 * 60.0 * 24
	TIME_7DAY     = 1.0 * 60.0 * 60.0 * 24 * 7
	TIME_30DAY    = 1.0 * 60.0 * 60.0 * 24 * 30
	TIME_KEEP     = -1
	TIME_KEEPN    = TIME_DAY * 365 * 100
	TIME_EXPIREDN = -1.0
)

func RandomRange(min int, max int) uint32 {
	if min > max {
		min = 0
	}

	rand.Seed(time.Now().UnixNano())

	var r = uint32(rand.Int63n(int64((max - min) + 1)))
	return r + uint32(min)
}

// TIMESTAMP
func GetTimeStamp() uint32 {
	return uint32(time.Now().Unix())
}

func GetTimeStamp64() uint64 {
	return uint64(math.Round(float64(time.Now().UnixMicro()) * 0.001))
}

func GetTimeStamp64M() uint64 {
	return uint64(time.Now().UnixMicro())
}

func CheckTimestamp64(prev uint64, next uint64) int32 {
	if prev == 0 {
		return 0
	}
	if int64(prev) == -1 {
		return TIME_KEEP
	}
	return int32(next - prev)
}

func ExpiredTimestamp64(timestamp uint64, expired float32) float32 {
	if timestamp == 0 {
		return TIME_EXPIREDN
	}
	if int64(timestamp) == -1 {
		return TIME_KEEPN
	}

	t := timestamp + uint64(math.Round(float64(expired*1000.0)))
	n := CheckTimestamp64(GetTimeStamp64(), t)
	if n == 0 {
		return TIME_EXPIREDN
	}
	if n == -1 {
		return TIME_KEEPN
	}
	v := float32(n) * 0.001
	return v
}

// DATE
func DateFormat(date time.Time, level int) string {

	//
	if level == 0 {
		return date.Format("2006-01-02 15:04:05")
	} else if level == 1 {
		return date.Format("2006-01-02")
	} else if level == 2 {
		return date.Format("15:04:05")
	} else if level == 3 {
		return date.Format("2006-01-02 15:04:05.000")
	} else if level == 8 {
		return date.Format("2006-01-02 15:04:05.000 -0700 MST")
	} else if level == 9 {
		return date.Format("2006-01-02 15:04:05.000 Mon")
	}

	//yyyy-MM-dd HH:mm:ss.ms ZONE MST WEEK
	return date.Format("2006-01-02 15:04:05.000 -0700 MST Mon")
}

func CheckAccountIDX(idx string, min int, max int) bool {
	regex := regexp.MustCompile("^[0-9]+$")
	if !regex.MatchString(idx) {
		return false
	}

	l := len(idx)
	return l >= min && l <= max
}

// default : 8 number
// level 1: 10 number
// level 2: 12 number
func GenerateIDX(level int) int64 {
	var date = time.Now()
	year := date.Year()%1000 + 1000
	a := int(date.Month())*10 + date.Day()
	b := date.Hour()
	c := date.Minute()
	x := int64(RandomRange(1000, 9999))
	var value int64 = int64(year + a)
	if level == 1 {
		value = value*100 + int64(b+c)
	} else if level == 2 {
		y := int64(RandomRange(1000, 9999))
		value = value*100 + y
	} else if level == 3 || level == 4 {
		value = value + int64(b)
	}

	value = value*1000 + x

	var cc = []int{1, 2, 3, 4, 5, 6, 7}
	var n = value
	i := 0
	v := 0
	for n > 0 {
		var m = int(n % 10)
		v = v + m*cc[i%len(cc)]
		n = n / 10
		i++
	}

	value = value*10 + int64(v%10)
	var cx = []int{30, 31, 32, 33, 35, 36, 38, 39}
	var cv = int64(cx[v%len(cx)])
	if level == 3 || level == 4 {
		value = cv*10000*10000 + value
	}
	if level == 4 {
		value = 100*10000*10000 + value
	}
	return value
}
