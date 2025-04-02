package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"go-cron/internal/config"
	"go-cron/internal/environment"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
	mu      sync.Mutex
	logFile *os.File
}

var (
	instance *Logger
	once     sync.Once
	logDir   = "./logs"
)

func Instance(level ...string) *Logger {
	once.Do(func() {
		instance = newLogger(level...)
	})
	return instance
}

func Shutdown() {
	if instance != nil && instance.logFile != nil {
		instance.logFile.Close()
		instance.logFile = nil
	}
}

func SetLogLevel(level string) {
	if instance == nil {
		return
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()
	setLogLevel(instance.Logger, level)
}

func UpdateFromConfig(cfg *config.Config) {
	if cfg != nil && cfg.Server.LogLevel != "" {
		SetLogLevel(cfg.Server.LogLevel)
	}
}

func (l *Logger) WithField(key string, value any) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

func (l *Logger) WithFields(fields map[string]any) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

func newLogger(level ...string) *Logger {
	logger := logrus.New()

	l := &Logger{
		Logger:  logger,
		logFile: nil,
	}

	configureFormatter(logger)
	configureLogLevel(logger, level...)
	configureLogOutput(l)

	return l
}

func configureFormatter(logger *logrus.Logger) {
	logger.SetReportCaller(true)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat:   time.RFC3339Nano,
		DisableHTMLEscape: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			functionName := filepath.Base(f.Function)
			fileName := filepath.Base(f.File)
			return functionName, fmt.Sprintf("%s:%d", fileName, f.Line)
		},
	})
}

func configureLogLevel(logger *logrus.Logger, level ...string) {
	if len(level) > 0 && level[0] != "" {
		setLogLevel(logger, level[0])
		return
	}

	cfg, err := config.Instance()
	if config.IsValid(cfg, err) && cfg.Server.LogLevel != "" {
		setLogLevel(logger, cfg.Server.LogLevel)
		return
	}

	if config.IsValid(cfg, err) {
		logger.SetLevel(environmentLevel(cfg.Server.Env))
		return
	}

	logger.SetLevel(logrus.InfoLevel)
}

func setLogLevel(logger *logrus.Logger, level string) {
	levelMap := map[string]logrus.Level{
		"panic": logrus.PanicLevel,
		"fatal": logrus.FatalLevel,
		"error": logrus.ErrorLevel,
		"warn":  logrus.WarnLevel,
		"info":  logrus.InfoLevel,
		"debug": logrus.DebugLevel,
		"trace": logrus.TraceLevel,
	}

	if lvl, exists := levelMap[level]; exists {
		logger.SetLevel(lvl)
	} else {
		logger.WithField("level", level).Warn("Unknown log level specified, using default (info)")
		logger.SetLevel(logrus.InfoLevel)
	}
}

func environmentLevel(env environment.Environment) logrus.Level {
	if env == "" {
		return logrus.InfoLevel
	}

	switch {
	case environment.IsProduction(env):
		return logrus.WarnLevel
	case environment.IsStaging(env):
		return logrus.WarnLevel
	case environment.IsTesting(env):
		return logrus.InfoLevel
	case environment.IsDevelopment(env):
		return logrus.DebugLevel
	default:
		return logrus.InfoLevel
	}
}

func configureLogOutput(l *Logger) {
	var writers []io.Writer

	cfg, err := config.Instance()
	consoleOutput := true
	fileOutput := true

	if config.IsValid(cfg, err) {
		consoleOutput = cfg.Server.LogOutput.Console
		fileOutput = cfg.Server.LogOutput.File
	}

	if consoleOutput {
		writers = append(writers, os.Stdout)
	}

	if fileOutput {
		if err := setupFileOutput(l, &writers); err != nil {
			l.WithError(err).Error("Failed to setup file output")
		}
	}

	switch len(writers) {
	case 0:
		l.SetOutput(os.Stdout)
	case 1:
		l.SetOutput(writers[0])
	default:
		l.SetOutput(io.MultiWriter(writers...))
	}

	l.Info("Logger initialized successfully")
}

func setupFileOutput(l *Logger, writers *[]io.Writer) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		l.WithError(err).Error("Failed to create log directory")
		return err
	}

	logFileName := filepath.Join(logDir, "go-cron.log")
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		l.WithError(err).Error("Failed to create log file")
		return err
	}

	l.logFile = file
	*writers = append(*writers, file)

	return nil
}
