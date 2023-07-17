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
	"strings"
	"sync"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

var md goldmark.Markdown
var contentDir string
var templateDir string
var staticDir string
var outDir string

func main() {
	md = goldmark.New(
		goldmark.WithExtensions(
			&frontmatter.Extender{},
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

	pagePaths, err := filepath.Glob(contentDir + "/**.md")
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
			pages[i] = ParsePage(path)
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

	staticPaths, err := filepath.Glob(staticDir + "/**")
	if err != nil {
		panic(err)
	}
	for _, path := range staticPaths {
		fmt.Printf("Copy %s... ", path)
		src, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer src.Close()
		outPath := outDir + strings.TrimPrefix(path, staticDir)
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
	}
}

type PageMeta struct {
	Title string `yaml:"title"`
	Desc  string `yaml:"description"`
	Type  string `yaml:"type"`
}

type RenderContext struct {
	Path    string
	Content template.HTML
	Meta    PageMeta
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

	return RenderContext{url, template.HTML(content), meta}
}

func RenderPage(ctx RenderContext) {
	outPath := outDir + ctx.Path + ".html"
	outDir := path.Dir(outPath)

	fmt.Printf("Render %s\n", outPath)

	tmpl, err := template.ParseFiles(templateDir+"/main.html", templateDir+"/types/"+ctx.Meta.Type+".html")
	if err != nil {
		panic(err)
	}

	os.MkdirAll(outDir, os.ModePerm)
	f, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	tmpl.Execute(f, ctx)
}
