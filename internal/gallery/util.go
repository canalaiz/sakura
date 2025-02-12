package gallery

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "path/filepath"
)

func autoSenseJSON(content []byte) (string, error) {
    var posts []MediaContainer
    if err := json.Unmarshal(content, &posts); err == nil {
        return "posts", nil
    }

    var archived struct {
        Media []MediaContainer `json:"ig_archived_post_media"`
    }
    if err := json.Unmarshal(content, &archived); err == nil {
        return "archived", nil
    }

    var reels struct {
        Media []MediaContainer `json:"ig_reels_media"`
    }
    if err := json.Unmarshal(content, &reels); err == nil {
        return "reels", nil
    }

    var stories struct {
        Media []Media `json:"ig_stories"`
    }
    if err := json.Unmarshal(content, &stories); err == nil {
        return "stories", nil
    }

    var igtv struct {
        Media []MediaContainer `json:"ig_igtv_media"`
    }
    if err := json.Unmarshal(content, &igtv); err == nil {
        return "igtv", nil
    }

    var other struct {
        Media []MediaContainer `json:"ig_other_media"`
    }
    if err := json.Unmarshal(content, &other); err == nil {
        return "other", nil
    }

    return "", fmt.Errorf("unknown JSON format")
}

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
