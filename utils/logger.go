package utils

type LogLogger struct {
}

var Logger *LogLogger = nil

func NewLogger() *LogLogger {
	return &LogLogger{}
}

func (self *LogLogger) Init() {
	Logger = self

	LogInit()
}

func (logger LogLogger) LogDebug(args ...interface{}) {
	LogDebug(args...)
}

func (logger LogLogger) Log(args ...interface{}) {
	Log(args...)
}

func (logger LogLogger) LogWarning(args ...interface{}) {
	LogWarn(args...)
}

func (logger LogLogger) LogError(args ...interface{}) {
	LogError(args...)
}
