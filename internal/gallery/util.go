package gallery

import (
    "log"
    "os"
    "path/filepath"
)

func GetAbsolutePath(path string) string {
    absPath, err := filepath.Abs(path)
    if err != nil {
        LogWarn("Warning: Unable to determine absolute path for %s, using provided path.\n", path)
        return path
    }
    return absPath
}

func CreateOutputDir(path string) {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        err := os.MkdirAll(path, 0755)
        if err != nil {
            LogFatal("Failed to create output directory %s: %v", path, err)
        }
    }
}

const (
    LOG_LEVEL_QUIET = iota
    LOG_LEVEL_NORMAL
    LOG_LEVEL_VERBOSE
)

var logLevel = LOG_LEVEL_NORMAL

func SetLogLevel(level int) {
    logLevel = level
}

func LogInfo(format string, v ...interface{}) {
    if logLevel >= LOG_LEVEL_NORMAL {
        log.Printf(format, v...)
    }
}

func LogWarn(format string, v ...interface{}) {
    if logLevel >= LOG_LEVEL_NORMAL {
        log.Printf(format, v...)
    }
}

func LogVerbose(format string, v ...interface{}) {
    if logLevel >= LOG_LEVEL_VERBOSE {
        log.Printf(format, v...)
    }
}

func LogFatal(format string, v ...interface{}) {
    log.Fatalf(format, v...)
}
