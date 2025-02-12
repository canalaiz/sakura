package gallery

import (
    "embed"
    "encoding/json"
    "fmt"
    "html/template"
    "io"
    "os"
    "path/filepath"
)

//go:embed templates/*.html
var embeddedTemplates embed.FS

type Media struct {
    URI   string `json:"uri"`
    Title string `json:"title"`
}

type MediaContainer struct {
    Media []Media `json:"media"`
}

type PageData struct {
    Media     []Media
    Prev      string
    Next      string
    Title     string
    PrevLabel string
    NextLabel string
    PageNum   int
}

func readMedia(inputDir string) ([]Media, error) {
    var mediaItems []Media
    files, err := os.ReadDir(inputDir)
    if err != nil {
        LogWarn("Failed to read directory %s: %v", inputDir, err)
        return nil, err
    }

    for _, file := range files {
        if filepath.Ext(file.Name()) == ".json" {
            filePath := filepath.Join(inputDir, file.Name())
            content, err := os.ReadFile(filePath)
            if err != nil {
                LogWarn("Failed to read file %s: %v", filePath, err)
                continue
            }

            fileType, err := autoSenseJSON(content)
            if err != nil {
                LogWarn("Failed to determine JSON type for file %s: %v", filePath, err)
                continue
            }

            switch fileType {
            case "posts":
                var mediaContainers []MediaContainer
                if err := json.Unmarshal(content, &mediaContainers); err != nil {
                    LogWarn("Failed to unmarshal JSON from file %s: %v", filePath, err)
                    continue
                }
                for _, mediaContainer := range mediaContainers {
                    for _, media := range mediaContainer.Media {
                        mediaItems = append(mediaItems, media)
                    }
                }

            case "archived":
                var archived struct {
                    Media []MediaContainer `json:"ig_archived_post_media"`
                }
                if err := json.Unmarshal(content, &archived); err != nil {
                    LogWarn("Failed to unmarshal JSON from file %s: %v", filePath, err)
                    continue
                }
                for _, mediaContainer := range archived.Media {
                    for _, media := range mediaContainer.Media {
                        mediaItems = append(mediaItems, media)
                    }
                }

            case "reels":
                var reels struct {
                    Media []MediaContainer `json:"ig_reels_media"`
                }
                if err := json.Unmarshal(content, &reels); err != nil {
                    LogWarn("Failed to unmarshal JSON from file %s: %v", filePath, err)
                    continue
                }
                for _, mediaContainer := range reels.Media {
                    for _, media := range mediaContainer.Media {
                        mediaItems = append(mediaItems, media)
                    }
                }

            case "stories":
                var stories struct {
                    Media []Media `json:"ig_stories"`
                }
                if err := json.Unmarshal(content, &stories); err != nil {
                    LogWarn("Failed to unmarshal JSON from file %s: %v", filePath, err)
                    continue
                }
                for _, media := range stories.Media {
                    mediaItems = append(mediaItems, media)
                }

            case "igtv":
                var igtv struct {
                    Media []MediaContainer `json:"ig_igtv_media"`
                }
                if err := json.Unmarshal(content, &igtv); err != nil {
                    LogWarn("Failed to unmarshal JSON from file %s: %v", filePath, err)
                    continue
                }
                for _, mediaContainer := range igtv.Media {
                    for _, media := range mediaContainer.Media {
                        mediaItems = append(mediaItems, media)
                    }
                }

            case "other":
                var other struct {
                    Media []MediaContainer `json:"ig_other_media"`
                }
                if err := json.Unmarshal(content, &other); err != nil {
                    LogWarn("Failed to unmarshal JSON from file %s: %v", filePath, err)
                    continue
                }
                for _, mediaContainer := range other.Media {
                    for _, media := range mediaContainer.Media {
                        mediaItems = append(mediaItems, media)
                    }
                }

            default:
                LogWarn("Unknown JSON type for file %s", filePath)
            }
        }
    }

    LogInfo("Read %d media items from %s", len(mediaItems), inputDir)
    LogVerbose("Media: %+v", mediaItems)

    return mediaItems, nil
}

func loadTemplates(templateDir string) (*template.Template, error) {
    if templateDir == "" {
        LogInfo("Loading embedded templates")
        return template.ParseFS(embeddedTemplates, "templates/*.html")
    }
    LogInfo("Loading templates from %s", templateDir)
    return template.ParseGlob(filepath.Join(templateDir, "*.html"))
}

func copyMedia(src, dst string) error {
    srcFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    dstFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer dstFile.Close()

    _, err = io.Copy(dstFile, srcFile)
    return err
}

func Generate(inputDir, outputDir, templateDir string, mediaPerPage int, prevLabel, nextLabel string) {
    LogInfo("Generating gallery with inputDir=%s, outputDir=%s, templateDir=%s, mediaPerPage=%d, prevLabel=%s, nextLabel=%s",
        inputDir, outputDir, templateDir, mediaPerPage, prevLabel, nextLabel)

    templates, err := loadTemplates(templateDir)
    if err != nil {
        LogFatal("Failed to parse templates: %v", err)
    }

    mediaItems, err := readMedia(inputDir)
    if err != nil {
        LogFatal("Failed to read media items: %v", err)
    }

    numPages := (len(mediaItems) + mediaPerPage - 1) / mediaPerPage

    for pageNum := 1; pageNum <= numPages; pageNum++ {
        start := (pageNum - 1) * mediaPerPage
        end := pageNum * mediaPerPage
        if end > len(mediaItems) {
            end = len(mediaItems)
        }

        pageMediaOutputDir := filepath.Join(outputDir, fmt.Sprintf("media_page_%d", pageNum))
        if err := os.MkdirAll(pageMediaOutputDir, 0755); err != nil {
            LogFatal("Failed to create media directory for page %d: %v", pageNum, err)
        }

        for i, media := range mediaItems[start:end] {
            LogVerbose("Copying media from %s with inputDir=%s", media.URI, inputDir)
            srcPath := filepath.Join(inputDir, "../../", media.URI)
            dstPath := filepath.Join(pageMediaOutputDir, filepath.Base(media.URI))
            if err := copyMedia(srcPath, dstPath); err != nil {
                LogWarn("Failed to copy media %s to %s: %v", srcPath, dstPath, err)
            }
            mediaItems[start+i].URI = filepath.ToSlash(filepath.Join(fmt.Sprintf("media_page_%d", pageNum), filepath.Base(media.URI)))
        }

        pageData := PageData{
            Media:     mediaItems[start:end],
            Prev:      "",
            Next:      "",
            Title:     fmt.Sprintf("Sakura Gallery Page %d", pageNum),
            PrevLabel: prevLabel,
            NextLabel: nextLabel,
            PageNum:   pageNum,
        }
        if pageNum > 1 {
            pageData.Prev = fmt.Sprintf("gallery_page_%d.html", pageNum-1)
        }
        if pageNum < numPages {
            pageData.Next = fmt.Sprintf("gallery_page_%d.html", pageNum+1)
        }

        outputFile, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("gallery_page_%d.html", pageNum)))
        if err != nil {
            LogFatal("Failed to create output file: %v", err)
        }
        defer outputFile.Close()

        if err := templates.ExecuteTemplate(outputFile, "base", pageData); err != nil {
            LogFatal("Failed to execute base template: %v", err)
        }

        LogVerbose("Gallery page generated at %s/gallery_page_%d.html", outputDir, pageNum)
    }
}
