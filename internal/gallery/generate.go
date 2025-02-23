package gallery

import (
    "embed"
    "fmt"
    "html/template"
    "io"
    "os"
    "path/filepath"
)

//go:embed templates/*.html
var embeddedTemplates embed.FS

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
    var mediaList []Media
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

            mediaItems, _, err := autoSenseContents(content)
            if err != nil {
                LogWarn("Failed to determine JSON type for file %s: %v", filePath, err)
                continue
            }

            mediaList = append(mediaList, mediaItems...)
        }
    }

    LogInfo("Read %d media from %s", len(mediaList), inputDir)
    LogVerbose("Media: %+v", mediaList)

    return mediaList, nil
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
            LogVerbose("Media: URI=%s, Type=%s, CreatedAt=%d", media.URI, media.Type, media.CreatedAt)
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
