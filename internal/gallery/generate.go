package gallery

import (
    "embed"
    "fmt"
    "encoding/base64"
    "html/template"
    "io"
    "os"
    "path/filepath"
    "bytes"
    "io/ioutil"
    "golang.org/x/text/encoding/charmap"
    "golang.org/x/text/transform"    
    "strings"
)

//go:embed templates/*.html
var embeddedTemplates embed.FS

type PageData struct {
    MediaContainers     []MediaContainer
    Prev      string
    Next      string
    Title     string
    PrevLabel string
    NextLabel string
    PageNum   int
}

func base64Encode(s string) string {
    return base64.StdEncoding.EncodeToString([]byte(s))
}

func isVideoUri(uri string) bool {
	if i := strings.LastIndex(uri, "/"); i >= 0 {
		uri = uri[i+1:]
	}
	if dot := strings.LastIndex(uri, "."); dot >= 0 {
		ext := strings.ToLower(uri[dot:])
		switch ext {
		case ".mp4", ".webm", ".ogg", ".ogv":
			return true
		default:
			return false
		}
	}
	return true
}

// fixes mojibake by converting characters outside the ASCII range to their byte representation
func fixMojibake(s string) string {
    b := make([]byte, 0, len(s))
    for _, r := range s {
        if r > 255 {
            b = append(b, []byte(string(r))...)
        } else {
            b = append(b, byte(r))
        }
    }
    return string(b)
}


func repairEncoding(s string) string {
    reader := transform.NewReader(bytes.NewReader([]byte(s)), charmap.Windows1252.NewDecoder())
    fixed, err := ioutil.ReadAll(reader)
    if err != nil {
        return s
    }
    return string(fixed)
}

func readMedia(inputDir string) ([]MediaContainer, error) {
    var mediaContainers []MediaContainer
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

            // Fix titles
            for i, container := range mediaItems {
                container.Title = fixMojibake(container.Title)
                for j, media := range container.Media {
                    media.Title = fixMojibake(media.Title)
                    container.Media[j] = media
                }
                mediaItems[i] = container
            }

            mediaContainers = append(mediaContainers, mediaItems...)
        }
    }

    LogInfo("Read %d media containers from %s", len(mediaContainers), inputDir)
    LogVerbose("Media Containers: %+v", mediaContainers)

    return mediaContainers, nil
}

func loadTemplates(templateDir string) (*template.Template, error) {
    // Inject these functions on templates
    funcMap := template.FuncMap{
        "base64Encode": base64Encode,
        "isVideoUri": isVideoUri,
    }

    var tmpl *template.Template

    if templateDir == "" {
        LogInfo("Loading embedded templates")
        tmpl = template.New("").Funcs(funcMap)
        return tmpl.ParseFS(embeddedTemplates, "templates/*.html")
    }

    LogInfo("Loading templates from %s", templateDir)
    tmpl = template.New("").Funcs(funcMap)
    return tmpl.ParseGlob(filepath.Join(templateDir, "*.html"))
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

    mediaContainers, err := readMedia(inputDir)
    if err != nil {
        LogFatal("Failed to read media: %v", err)
    }

    numPages := (len(mediaContainers) + mediaPerPage - 1) / mediaPerPage

    for pageNum := 1; pageNum <= numPages; pageNum++ {
        start := (pageNum - 1) * mediaPerPage
        end := pageNum * mediaPerPage
        if end > len(mediaContainers) {
            end = len(mediaContainers)
        }

        pageMediaOutputDir := filepath.Join(outputDir, fmt.Sprintf("media_page_%d", pageNum))
        if err := os.MkdirAll(pageMediaOutputDir, 0755); err != nil {
            LogFatal("Failed to create media directory for page %d: %v", pageNum, err)
        }

        for _, mediaContainer := range mediaContainers[start:end] {
            for i, media := range mediaContainer.Media {
                LogVerbose("Copying media from %s with inputDir=%s", media.URI, inputDir)
                srcPath := filepath.Join(inputDir, "../../", media.URI)
                dstPath := filepath.Join(pageMediaOutputDir, filepath.Base(media.URI))
                if err := copyMedia(srcPath, dstPath); err != nil {
                    LogWarn("Failed to copy media %s to %s: %v", srcPath, dstPath, err)
                }
                mediaContainer.Media[i].URI = filepath.ToSlash(filepath.Join(fmt.Sprintf("media_page_%d", pageNum), filepath.Base(media.URI)))
            }
        }

        pageData := PageData{
            MediaContainers: mediaContainers[start:end],
            Prev:            "",
            Next:            "",
            Title:           fmt.Sprintf("Sakura Gallery Page %d", pageNum),
            PrevLabel:       prevLabel,
            NextLabel:       nextLabel,
            PageNum:         pageNum,
        }
        if pageNum > 1 {
            if pageNum - 1 == 1 {
                pageData.Prev = "index.html"
            } else {
                pageData.Prev = fmt.Sprintf("index_%d.html", pageNum-1)
            }
        }
        if pageNum < numPages {
            pageData.Next = fmt.Sprintf("index_%d.html", pageNum+1)
        }

        var outputFilename string
        if pageNum == 1 {
            outputFilename = "index.html"
        } else {
            outputFilename = fmt.Sprintf("index_%d.html", pageNum)
        }

        outputFile, err := os.Create(filepath.Join(outputDir, outputFilename))
        if err != nil {
            LogFatal("Failed to create output file: %v", err)
        }
        defer outputFile.Close()

        if err := templates.ExecuteTemplate(outputFile, "base", pageData); err != nil {
            LogFatal("Failed to execute base template: %v", err)
        }

        LogVerbose("Gallery page generated at %s/%s", outputDir, outputFilename)
    }
}