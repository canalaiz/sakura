# Sakura

Sakura is a command line tool to generate static HTML photo galleries from downloaded Instagram photos.

## Table of Contents
- [Getting Started](#getting-started)
- [Installation](#installation)
- [Usage](#usage)
  - [Serve Command](#serve-command)
  - [Templates](#templates)  
- [Flags](#flags)
- [Examples](#examples)
- [License](#license)

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

## Installation

First, ensure you have [Go](https://golang.org/doc/install) installed on your machine. Sakura requires Go version 1.23.6 or later.

To install Sakura, clone this repository and build the project:

```bash
git clone https://github.com/canalaiz/sakura.git
cd sakura
go build -o sakura
```

## Usage

Sakura provides a simple way to generate HTML photo galleries from Instagram JSON backup files.

### Serve Command

The `serve` command is used to generate the photo gallery. You can specify various options to customize the output.

### Templates

Sakura uses standard Go templates to generate HTML photo galleries. The binary includes default templates, but you can customize these by providing your own templates using the `--template` flag.

The repository includes a `templates` directory that provides examples of the default templates and the data passed to each page. This can serve as a guide for creating your own custom templates.

#### Customizing Templates

Templates in Go can be composed of multiple files, and Sakura supports this feature. You can create a custom template directory with multiple template files, each containing different parts of the HTML structure. All template files must be placed at the root level of the specified templates directory.

By default, the following templates are used:

- `base.html`: The main template for generating gallery pages.
- `photo.html`: A template for individual photo entries.

You can modify these templates or create new ones based on your requirements. Simply provide the path to your custom templates directory using the `--template` flag.

Example of the default `base.html` template:
```html
{{ define "base" }}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }}</title>
    <style>
        /* Add your CSS styles here */
    </style>
</head>
<body>
    <div class="grid-container">
        {{ template "content" .Photos }}
    </div>
    <div class="pagination">
        {{if .Prev}}
        <a href="{{.Prev}}">{{ .PrevLabel }}</a>
        {{end}}
        {{if .Next}}
        <a href="{{.Next}}">{{ .NextLabel }}</a>
        {{end}}
    </div>
</body>
</html>
{{ end }}
```

#### Data Passed to Templates

Each template receives a set of data that can be used to generate the HTML content. The main data structure passed to the templates includes:

- `Title`: The title of the gallery page.
- `Photos`: A list of photos to be displayed on the page.
- `Prev`: The link to the previous gallery page.
- `Next`: The link to the next gallery page.
- `PrevLabel`: The label for the 'Previous' button.
- `NextLabel`: The label for the 'Next' button.
- `PageNum`: The current page number of the gallery.

You can use these fields in your custom templates to display the relevant information on your gallery pages.

For more details on Go templates, refer to the [Go template documentation](https://golang.org/pkg/text/template/).

#### Flags

- `--dir, -d` : Path to the Instagram backup root folder (default ".")
- `--output, -o` : Path to the output folder for HTML pages (default ".")
- `--template, -t` : Path to the template folder (default "")
- `--photos-per-page, -p` : Number of photos per page (default 8)
- `--title, -T` : Title of the gallery (default "Sakura Gallery Page")
- `--prev-label, -P` : Label for the 'Previous' button (default "Previous")
- `--next-label, -N` : Label for the 'Next' button (default "Next")
- `--verbose, -v` : Enable verbose logging
- `--quiet, -q` : Enable quiet mode (no logging)

## Examples

### Basic Usage

Generate a photo gallery from the current directory:

```bash
./sakura serve --dir . --output ./output --photos-per-page 4 --prev-label "Indietro" --next-label "Avanti"
```

### Custom Template Usage

Generate a photo gallery from the current directory:

```bash
./sakura serve --dir . --output ./output --template ./custom_templates --photos-per-page 4 --prev-label "Indietro" --next-label "Avanti"
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
