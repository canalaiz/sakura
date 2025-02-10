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
    Photos    []Photo
    Prev      string
    Next      string
    Title     string
    PrevLabel string
    NextLabel string
    PageNum   int
}

type Photo struct {
    Path  string `json:"path"`
    Title string `json:"title"`
}

func readPhotos(inputDir string) ([]Photo, error) {
    var photos []Photo
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
                return nil, err
            }

            var mediaContainers []MediaContainer
            if err := json.Unmarshal(content, &mediaContainers); err != nil {
                LogWarn("Failed to unmarshal JSON from file %s: %v", filePath, err)
                return nil, err
            }

            for _, mediaContainer := range mediaContainers {
                for _, media := range mediaContainer.Media {
                    photo := Photo{
                        Path:  media.URI,
                        Title: media.Title,
                    }
                    photos = append(photos, photo)
                }
            }
        }
    }

    LogInfo("Read %d photos from %s", len(photos), inputDir)
    LogVerbose("Photos: %+v", photos)

    return photos, nil
}

func loadTemplates(templateDir string) (*template.Template, error) {
    if templateDir == "" {
        LogInfo("Loading embedded templates")
        return template.ParseFS(embeddedTemplates, "templates/*.html")
    }
    LogInfo("Loading templates from %s", templateDir)
    return template.ParseGlob(filepath.Join(templateDir, "*.html"))
}

func copyPhoto(src, dst string) error {
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

func Generate(inputDir, outputDir, templateDir string, photosPerPage int, prevLabel, nextLabel string) {
    LogInfo("Generating gallery with inputDir=%s, outputDir=%s, templateDir=%s, photosPerPage=%d, prevLabel=%s, nextLabel=%s",
        inputDir, outputDir, templateDir, photosPerPage, prevLabel, nextLabel)

    templates, err := loadTemplates(templateDir)
    if err != nil {
        LogFatal("Failed to parse templates: %v", err)
    }

    photos, err := readPhotos(inputDir)
    if err != nil {
        LogFatal("Failed to read photos: %v", err)
    }

    numPages := (len(photos) + photosPerPage - 1) / photosPerPage

    for pageNum := 1; pageNum <= numPages; pageNum++ {
        start := (pageNum - 1) * photosPerPage
        end := pageNum * photosPerPage
        if end > len(photos) {
            end = len(photos)
        }

        pagePhotoOutputDir := filepath.Join(outputDir, fmt.Sprintf("photos_page_%d", pageNum))
        if err := os.MkdirAll(pagePhotoOutputDir, 0755); err != nil {
            LogFatal("Failed to create photos directory for page %d: %v", pageNum, err)
        }

        for i, photo := range photos[start:end] {
            LogVerbose("Copying photo from %s with inputDir=%s", photo.Path, inputDir)
            srcPath := filepath.Join(inputDir, "../../", photo.Path)
            dstPath := filepath.Join(pagePhotoOutputDir, filepath.Base(photo.Path))
            if err := copyPhoto(srcPath, dstPath); err != nil {
                LogWarn("Failed to copy photo %s to %s: %v", srcPath, dstPath, err)
            }
            photos[start+i].Path = filepath.Join(fmt.Sprintf("photos_page_%d", pageNum), filepath.Base(photo.Path))
        }

        pageData := PageData{
            Photos:    photos[start:end],
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
