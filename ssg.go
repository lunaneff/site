package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	chromaHtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkHtml "github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
)

var md goldmark.Markdown
var contentDir string
var templateDir string
var staticDir string
var outDir string

func main() {
	md = goldmark.New(
		goldmark.WithRendererOptions(
			goldmarkHtml.WithUnsafe(),
		),
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.Footnote,
			&frontmatter.Extender{},
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(
					chromaHtml.WithClasses(true),
					chromaHtml.TabWidth(2),
				),
			),
		),
	)

	flag.StringVar(&contentDir, "contentDir", "content", "the directory containing content files")
	flag.StringVar(&templateDir, "templateDir", "templates", "the directory containing template files")
	flag.StringVar(&staticDir, "staticDir", "static", "the directory containing static files which are copied to the output")
	flag.StringVar(&outDir, "out", "out", "the directory where files are generated")

	flag.Parse()

	if _, err := os.Stat(outDir); err == nil {
		fmt.Println("Clean output directory")
		os.RemoveAll(outDir)
	}

	var pagePaths []string
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".md") {
			pagePaths = append(pagePaths, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	numPages := len(pagePaths)

	var wg sync.WaitGroup
	pages := make([]RenderContext, numPages)
	for i, path := range pagePaths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			ctx := ParsePage(path)
			ctx.AllPages = pages
			pages[i] = ctx
		}(i, path)
	}
	wg.Wait()

	for _, ctx := range pages {
		wg.Add(1)
		go func(ctx RenderContext) {
			defer wg.Done()
			RenderPage(ctx)
		}(ctx)
	}
	wg.Wait()

	err = filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if path == staticDir {
			// this function is called for the dir itself; don't do anything in that case to avoid confusing output
			return nil
		}

		fmt.Printf("Copy %s... ", path)
		outPath := outDir + strings.TrimPrefix(path, staticDir)

		if info.IsDir() {
			err = os.MkdirAll(outPath, os.ModePerm)
			if err != nil {
				panic(err)
			}
			fmt.Println("Directory created")
			return nil
		}

		src, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer src.Close()
		dst, err := os.Create(outPath)
		if err != nil {
			panic(err)
		}
		defer dst.Close()

		num, err := io.Copy(dst, src)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%d bytes copied\n", num)

		return nil
	})
	if err != nil {
		panic(err)
	}
}

type PageMeta struct {
	Title     string         `yaml:"title"`
	Desc      string         `yaml:"description"`
	Type      string         `yaml:"type"`
	Published time.Time      `yaml:"published"`
	Custom    map[string]any `yaml:"custom"`
}

type RenderContext struct {
	Path     string
	Content  template.HTML
	Meta     PageMeta
	AllPages []RenderContext
}

func ParsePage(path string) RenderContext {
	fmt.Printf("Parse %s\n", path)

	src, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	ctx := parser.NewContext()
	var buf bytes.Buffer
	if err = md.Convert(src, &buf, parser.WithContext(ctx)); err != nil {
		panic(err)
	}

	d := frontmatter.Get(ctx)
	var meta PageMeta
	if err = d.Decode(&meta); err != nil {
		panic(err)
	}

	url := strings.TrimPrefix(strings.TrimSuffix(path, ".md"), contentDir)

	content := string(buf.Bytes())

	return RenderContext{url, template.HTML(content), meta, nil}
}

func RenderPage(ctx RenderContext) {
	outPath := outDir + ctx.Path + ".html"
	outDir := path.Dir(outPath)

	fmt.Printf("Render %s\n", outPath)

	tmpl := template.New("main")

	tmpl = tmpl.Funcs(template.FuncMap{
		"url": func(path string) string {
			return strings.TrimSuffix(path+".html", "index.html")
		},
		"filterType": func(neededType string, pages []RenderContext) []RenderContext {
			ret := make([]RenderContext, 0)
			for _, page := range pages {
				if page.Meta.Type == neededType {
					ret = append(ret, page)
				}
			}
			return ret
		},
		"formatTime": func(format string, time time.Time) string {
			return time.Format(format)
		},
		"mostRecent": func(pages []RenderContext) []RenderContext {
			pages = append([]RenderContext{}, pages...) // make a copy; sort mutates
			sort.Slice(pages, func(i, j int) bool {
				return pages[i].Meta.Published.After(pages[j].Meta.Published)
			})
			return pages
		},
	})

	tmpl, err := tmpl.ParseFiles(templateDir+"/main.html", templateDir+"/types/"+ctx.Meta.Type+".html")
	if err != nil {
		panic(err)
	}

	os.MkdirAll(outDir, os.ModePerm)
	f, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = tmpl.Execute(f, ctx)
	if err != nil {
		panic(err)
	}
}

func PageUrl(path string) string {
	return strings.TrimSuffix(path+".html", "index.html")
}
