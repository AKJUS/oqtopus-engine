package log

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/oqtopus-team/oqtopus-engine/coreapp/common"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/core"
	"go.uber.org/zap"
)

const MetricsLogTaskName = "metrics_log"
const queueLengthKeyInMetrics = "queue_length"

type MetricsLogTaskImpl struct {
	FileDir string `toml:"file_dir"`

	dl *dailyLogger
	sc *core.SystemComponents

	core.DefaultTaskImpl
}

func setupMetricsLogTask(fileDir string) (*dailyLogger, error) {
	if err := common.IsDirWritable(fileDir); err != nil {
		return nil, fmt.Errorf("failed to write to %s: %w", fileDir, err)
	}
	newDailyLogger := newDailyLogger(fileDir)
	slog.SetDefault(slog.New(slog.NewJSONHandler(newDailyLogger, nil)))
	return newDailyLogger, nil
}

func (m *MetricsLogTaskImpl) Setup() error {
	dl, err := setupMetricsLogTask(m.FileDir)
	if err != nil {
		zap.L().Error("failed to set up metrics log task", zap.Error(err))
		return err
	}
	sc := core.GetSystemComponents()
	m.dl = dl
	m.sc = sc
	return nil
}

func (m *MetricsLogTaskImpl) GetEmptyParams() interface{} {
	return m
}

func (m *MetricsLogTaskImpl) SetParams(p interface{}) error {
	if p == nil {
		msg := "no params for metrics log task"
		zap.L().Debug(msg)
		return nil
	}
	mp, ok := p.(map[string]interface{})
	if !ok {
		msg := fmt.Errorf("failed to set params for metrics log task/params: %s", p)
		zap.L().Error(msg.Error())
		return msg
	}
	if fileDir, ok := mp["file_dir"].(string); ok {
		m.FileDir = fileDir
	}
	return nil
}

func (m *MetricsLogTaskImpl) Task() {
	slog.Info(
		"Metrics",
		slog.Int(
			queueLengthKeyInMetrics,
			m.sc.GetCurrentQueueSize()),
	)
}

func (m *MetricsLogTaskImpl) Cleanup() {
	m.dl.Close()
}

type dailyLogger struct {
	mu              sync.Mutex
	fileDir         string
	currentFileName string
	file            *os.File
}

func newDailyLogger(fileDir string) *dailyLogger {
	return &dailyLogger{
		fileDir: fileDir,
	}
}

func (dl *dailyLogger) Write(p []byte) (n int, err error) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	fileName := fmt.Sprintf("metrics-%s.log", time.Now().Format("2006-01-02"))
	filePath := filepath.Join(dl.fileDir, fileName)
	currentFilePath := filepath.Join(dl.fileDir, dl.currentFileName)

	if dl.file == nil || currentFilePath != filePath {
		if dl.file != nil {
			dl.file.Close()
		}
		var err error
		dl.file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return 0, err
		}
		dl.currentFileName = fileName
	}

	return dl.file.Write(p)
}

func (dl *dailyLogger) Close() error {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	if dl.file != nil {
		return dl.file.Close()
	}
	return nil
}
