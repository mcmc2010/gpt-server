package utils

import (
	"math"
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
