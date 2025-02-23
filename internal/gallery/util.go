package gallery

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "path/filepath"
)

type Media struct {
    URI   string `json:"uri"`
    Title string `json:"title"`
    Type string `json:"type"`
    CreatedAt int64 `json:"creation_timestamp"`
}

type MediaContainer struct {
    Media []Media `json:"media"`
}

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

func autoSenseContents(content []byte) ([]Media, string, error) {
    var mediaList []Media

    var posts []MediaContainer
    if err := json.Unmarshal(content, &posts); err == nil {
        for _, mediaContainer := range posts {
            for _, media := range mediaContainer.Media {
                media.Type = "posts"
                mediaList = append(mediaList, media)
            }
        }
        return mediaList, "posts", nil
    }

    var archived struct {
        Media []MediaContainer `json:"ig_archived_post_media"`
    }
    if err := json.Unmarshal(content, &archived); err == nil {
        for _, mediaContainer := range archived.Media {
            for _, media := range mediaContainer.Media {
                media.Type = "archived"
                mediaList = append(mediaList, media)
            }
        }
        return mediaList, "archived", nil
    }

    var reels struct {
        Media []MediaContainer `json:"ig_reels_media"`
    }
    if err := json.Unmarshal(content, &reels); err == nil {
        for _, mediaContainer := range reels.Media {
            for _, media := range mediaContainer.Media {
                media.Type = "reels"
                mediaList = append(mediaList, media)
            }
        }
        return mediaList, "reels", nil
    }

    var stories struct {
        Media []struct {
            Media
        } `json:"ig_stories"`
    }
    if err := json.Unmarshal(content, &stories); err == nil {
        for _, media := range stories.Media {
            media.Media.Type = "stories"
            mediaList = append(mediaList, media.Media)
        }
        return mediaList, "stories", nil
    }

    var igtv struct {
        Media []MediaContainer `json:"ig_igtv_media"`
    }
    if err := json.Unmarshal(content, &igtv); err == nil {
        for _, mediaContainer := range igtv.Media {
            for _, media := range mediaContainer.Media {
                media.Type = "igtv"
                mediaList = append(mediaList, media)
            }
        }
        return mediaList, "igtv", nil
    }

    var other struct {
        Media []MediaContainer `json:"ig_other_media"`
    }
    if err := json.Unmarshal(content, &other); err == nil {
        for _, mediaContainer := range other.Media {
            for _, media := range mediaContainer.Media {
                media.Type = "other"
                mediaList = append(mediaList, media)
            }
        }
        return mediaList, "other", nil
    }

    return nil, "", fmt.Errorf("unknown JSON format")
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
