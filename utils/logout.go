package utils

import (
	"container/list"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	LogLevel_Debug    int = 1
	LogLevel_Info     int = 0
	LogLevel_Warnning int = 2
	LogLevel_Error    int = 3
	LogLevel_Max      int = 4
)

type LogItem struct {
	level        int
	name         string
	filename     string
	output_file  bool
	output_print bool
	output_time  bool
	lock         *sync.Mutex
	doing        bool //write log
	values       *list.List
}

type LogStats struct {
	status int
	eof    string
	cwd    string
	dir    string
	items  map[string]LogItem
}

var log_stats LogStats
var log_init_completed = false

func LogInit() bool {
	if log_init_completed {
		return true
	}

	log_stats.cwd = "."
	log_stats.eof = "\n"
	log_stats.dir = "logs"
	log_stats.status = -1

	//var date = time.Now()
	//log_filename = fmt.Sprintf("%s_%4d%02d%02d", name, date.Year(), date.Month(), date.Day())
	//log_filename = log_filename + ".log"

	var sys = runtime.GOOS
	if sys == "windows" {
		log_stats.eof = "\r\n"
	}

	cwd, _ := os.Getwd()
	log_stats.cwd = cwd
	log_stats.dir = cwd + "/" + "logs"

	_, err := os.Stat(log_stats.dir)
	if err != nil || os.IsNotExist(err) {
		err = os.MkdirAll(log_stats.dir, 0766)
		if err != nil {
			println("[Error] Make log dirs (%s) fail", log_stats.dir)
			return false
		}
	}

	log_stats.items = make(map[string]LogItem)
	log_stats.status = 0
	log_init_completed = true

	//
	LogAdd(LogLevel_Error, "Error", true, true)
	LogAdd(LogLevel_Warnning, "Warnning", true, true)
	LogAdd(LogLevel_Info, "Info", true, true)
	return true
}

func LogAdd(level int, name string, output_file bool, output_print bool) bool {
	if !log_init_completed {
		println("[Error] Add log (%s) fail", name)
		return false
	}

	var item LogItem
	item.output_time = true
	item.output_file = output_file
	item.output_print = output_print

	item.doing = false
	item.lock = &sync.Mutex{}
	item.values = nil

	item.level = level
	if item.level >= LogLevel_Max {
		item.level = LogLevel_Info
	}
	item.name = strings.ToLower(name)

	var date = time.Now()
	var filename = fmt.Sprintf("%s_%4d%02d%02d", item.name, date.Year(), date.Month(), date.Day())
	item.filename = filename + ".log"

	log_stats.items[item.name] = item
	return true
}

func log_output(name string, text string) {
	if !log_init_completed {
		return
	}

	var key = strings.TrimSpace(strings.ToLower(name))
	var item, ok = log_stats.items[key]
	if ok {
		item.lock.Lock()
		if item.values == nil {
			item.values = list.New()
		}
		item.values.PushBack(text)
		item.lock.Unlock()

		if item.lock.TryLock() {
			// Set list to temp, if doing is true.
			var lvalues *list.List = nil
			if !item.doing {
				item.doing = true
				lvalues = item.values
				item.values = nil
			}
			item.lock.Unlock()

			if lvalues != nil {
				for v := lvalues.Front(); v != nil; v = v.Next() {
					if item.output_print {
						log_print(item, v.Value.(string))
					}
					if item.output_file {
						log_file(item, v.Value.(string))
					}
				}
			}
			item.doing = false
		}
	}
}

func log_print(item LogItem, text string) {
	if item.level == LogLevel_Error {
		println("[ERROR] " + text)
	} else if item.level == LogLevel_Warnning {
		println("[WARNNING] " + text)
	} else if item.level == LogLevel_Debug {
		println("[DEBUG] " + text)
	} else {
		println("[INFO] " + text)
	}
}

func log_file(item LogItem, text string) bool {
	var fullname = fmt.Sprintf("logs/%s", item.filename)

	file, err := os.OpenFile(fullname, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		println("[Error] Append log (%s) fail.", item.filename)
		return false
	}

	var value = ""
	if item.level == LogLevel_Error {
		value = "[ERROR] " + text
	} else if item.level == LogLevel_Warnning {
		value = "[WARNNING] " + text
	} else if item.level == LogLevel_Debug {
		value = "[DEBUG] " + text
	} else {
		value = "[INFO] " + text
	}

	if item.output_time {
		var tm = time.Now()
		var ss = fmt.Sprintf("%02d:%02d:%02d", tm.Hour(), tm.Minute(), tm.Second())
		value = ss + " " + value
	}

	// Set file end, append line.
	file.Seek(0, os.SEEK_END)
	file.WriteString(value + log_stats.eof)
	file.Close()

	return true
}

func log_args(args ...interface{}) string {
	var text = ""
	values, ok := interface{}(args).([]interface{})
	if !ok {
		text = log_parse_args(args)
	} else {
		for _, v := range values {
			text = text + log_parse_args(v)
		}
	}
	return text
}

func log_parse_args(args interface{}) string {
	var text = ""
	switch args.(type) {
	case int8:
		text = strconv.Itoa(int(args.(int8)))
		break
	case uint8:
		text = strconv.Itoa(int(args.(uint8)))
		break
	case int16:
		text = strconv.Itoa(int(args.(int16)))
		break
	case uint16:
		text = strconv.Itoa(int(args.(uint16)))
		break
	case int:
		text = strconv.Itoa(int(args.(int)))
		break
	case uint:
		text = strconv.Itoa(int(args.(uint)))
		break
	case int32:
		text = strconv.Itoa(int(args.(int32)))
		break
	case uint32:
		text = strconv.Itoa(int(args.(uint32)))
		break
	case int64:
		text = strconv.FormatInt(args.(int64), 10)
		break
	case uint64:
		text = strconv.FormatUint(args.(uint64), 10)
		break
	case float32:
		text = strconv.FormatFloat(float64(args.(float32)), 'f', -1, 32)
		break
	case float64:
		text = strconv.FormatFloat(args.(float64), 'f', -1, 64)
		break
	case string:
		text = args.(string)
		break
	case interface{}:
		vv, ok := interface{}(args).([]interface{})
		if !ok {
			text = fmt.Sprintf("%+v", args) //log_parse_args(args)
		} else {
			for _, v := range vv {
				text = text + log_parse_args(v)
			}
		}
		break
	default:
		text = fmt.Sprintf("%+v", args)
		break
	}
	return text
}

func LogWithName(name string, args ...interface{}) {
	var values = log_args(args)
	log_output(name, values)
}

func Log(args ...interface{}) {
	var values = log_args(args)
	log_output("Info", values)
}

func LogDebug(args ...interface{}) {
	var values = log_args(args)
	log_output("Debug", values)
}

func LogWarn(args ...interface{}) {
	var values = log_args(args)
	log_output("Warnning", values)
}

func LogError(args ...interface{}) {
	var values = log_args(args)
	log_output("Error", values)
}
