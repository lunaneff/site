---
title: Hello World
description: It works! also this post describes how the static site generator works
type: post
published: 2023-07-18
---

This is the first ever post on my website, made with a fully custom static site generator! 

It's definitely not as powerful as something like [Hugo](https://gohugo.io), but when I tried
that one, I found it a bit too complicated to get it to work and the documentation was hard to
understand at times, so I decided to make my own.

If you want to use it for your own website, here's how:

## Templates

Templates are what defines the layout of your site. These are in a directory called `templates`
and are HTML documents with Go's template syntax which is documented [here](https://pkg.go.dev/text/template).
There's also a section further down which explains some of the basics of templates.

You need at least two template files, one called `main.html` and the other can have any name,
but it has to be in the `types` subdirectory and also have the HTML extension.

Inside `main.html`, you need to define a template called `main`. This template is called for each
page.

The template files in `types` can be structured however you like. Your `main.html` can access
templates defined in these.

A basic structure for these templates could look like this:

```html
<!-- main.html -->
{{ define "main" }}
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  {{ template "head" . }}
</head>
<body>
  <header><!-- ... --></header>
  <main>
    {{ template "content" . }}
  </main>
  <footer><!-- ... --></footer>
</body>
{{ end }}

<!-- page.html -->
{{ define "head" }}
<title>{{ .Meta.Title }}</title>
{{ end }}

{{ define "content" }}
<h1>{{ .Meta.Title }}</h1>
<p>{{ .Meta.Desc }}</p>
{{ .Content }}
{{ end }}
```

Then, you could add a Markdown file in your `content` directory, which looks like this:

```markdown
---
title: Test Page
description: This is a description
type: page
---

Markdown text goes here
```

## Template basics

Template directives are wrapped in double curly braces.

Some common directives:

- `{{ .What.Ever }}` - Replaced with the value of the given context variable. Will be escaped automatically.
- `{{ define "name" }}` - defines a new template with the given name
- `{{ template "name" context }}` - calls the given template, with the given context
- `{{ range .Slice }}` - Calls the block inside `range` once for each value in the slice, each time with the value as context
- `{{ end }}` - ends a block started with `define` or `range`
- `{{ $var = "value" }}` - Sets the given variable to the given value

There are also some functions, most are predefined by Go but my site generator adds some more:

- `url` - Converts a path as in the `Path` variable of the render context, to a relative-to-root URL including file extension
- `filterType type slice` - Filters the given slice of page contexts to only include the given type
- `formatTime format time` - Formats the given time value using the given format. Format is a date/time string referring to `2006-01-02T15:04:05-0700`
- `mostRecent` - Sorts the given slice of page contexts by publish date, descending

## The Context

Each template has a context, which can be any arbitrary value. In your `main` template,
this value is an instance of the `RenderContext` struct, as below:

```go
type RenderContext struct {
    Path     string          // Path to the page, without `content/` prefix or file extension
    Content  template.HTML   // Rendered content of the page
    Meta     PageMeta        // Frontmatter of the page
    AllPages []RenderContext // Render contexts of every page in the site, intended for page lists
}

type PageMeta struct {
    Title     string         `yaml:"title"`       // Title of the page
    Desc      string         `yaml:"description"` // Description of the page
    Type      string         `yaml:"type"`        // Type of the page. Used to determine which template in the `types` directory to load.
    Published time.Time      `yaml:"published"`   // Publish date/time of the page
    Custom    map[string]any `yaml:"custom"`      // Arbitrary data
}
```

## Content files

Content files are Markdown files with YAML or TOML-based frontmatter. YAML
frontmatter is wrapped in 3 hyphens, TOML in 3 pluses.

The main content is standard Markdown, with the following extensions:

### Tables

```markdown
| col1 | col2 |
|------|------|
| r1c1 | r2c2 |
| r2c1 | r2c2 |
```

| col1 | col2 |
|------|------|
| r1c1 | r2c2 |
| r2c1 | r2c2 |

### Strikethrough text

```markdown
Some of this text is ~~rendered as strikethrough~~
```

### Footnotes

```markdown
> To be or not to be, that is the question[^1]

You can also reference the same footnote twice[^1]

[^1]: William Shakespeare, Hamlet
```

> To be or not to be, that is the question[^1]

You can also reference the same footnote twice[^1]

[^1]: William Shakespeare, Hamlet

### Arbitrary HTML

```markdown
<div style="width: 200px; height: 200px; background: linear-gradient(to bottom right, #eb6f92, #f6c177);"></div>
```

<div style="width: 200px; height: 200px; background: linear-gradient(to bottom right, #eb6f92, #f6c177);"></div>

## Static files

Any files placed in the `static` directory will be copied over to the output
directory during the build. This can be used for CSS, images, JS, or anything
else that isn't content.
