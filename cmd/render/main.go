package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

const (
	outDir    = "www/"
	domain    = "routerd.net"
	githubOrg = "routerd"
)

// Add new go modules here and
// then execute `go run ./cmd/render`
var modules = []string{
	"kube-ipam",
	"go-firewalld",
}

var (
	moduleTemplateString = `<!DOCTYPE html>
<html>
<head>
	<meta http-equiv="content-type" content="text/html; charset=utf-8">
	<meta name="go-import" content="{{ .GoImport }}">
	<meta name="go-source" content="{{ .GoSource }}">
	<meta http-equiv="refresh" content="0; {{ .Redirect }}">
</head>
<body>
</body>
</html>`

	indexTemplateString = `<!DOCTYPE html>
<html>
<head>
	<meta http-equiv="content-type" content="text/html; charset=utf-8">
	<meta http-equiv="refresh" content="0; {{ .Redirect }}">
</head>
<body>
</body>
</html>`

	redirectsTemplateString = `
{{ range $module := .Modules -}}
/{{$module }}/* go-get=1 /{{$module }}/index.html 200
{{ end }}`
)

type moduleTemplateData struct {
	GoImport string
	GoSource string
	Redirect string
}

type indexTemplateData struct {
	Redirect string
}

type redirectsTemplateData struct {
	Modules []string
}

func main() {
	// Load templates
	moduleTemplate := template.Must(
		template.New("module page").Parse(moduleTemplateString),
	)
	indexTemplate := template.Must(
		template.New("index page").Parse(indexTemplateString),
	)
	redirectsTemplate := template.Must(
		template.New("netlify redirects config").Parse(redirectsTemplateString),
	)

	// Remove output folder
	err := os.RemoveAll(outDir)
	if err != nil {
		panic(fmt.Sprintf("Could not remove outDir: %s %v", outDir, err))
	}

	// Recreate output folder
	err = os.MkdirAll(outDir, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Could not recreate outDir: %s %v", outDir, err))
	}

	// Create CNAME file in output folder
	cnameFilePath := filepath.Join(outDir, "CNAME")
	file, err := os.Create(cnameFilePath)
	defer file.Close()
	if err != nil {
		panic(fmt.Sprintf("could not create cname file: %v", err))
	}
	fmt.Fprintln(file, domain)

	// Create subfolder and page for each module
	for _, module := range modules {
		subfolderPath := filepath.Join(outDir, module)
		pagePath := filepath.Join(subfolderPath, "index.html")

		// Create module subfolder
		err = os.Mkdir(subfolderPath, os.ModePerm)
		if err != nil {
			panic(fmt.Sprintf("Could not create subfolder for module: %s %s %v", module, subfolderPath, err))
		}

		// Create module page file
		file, err := os.Create(pagePath)
		defer file.Close()
		if err != nil {
			panic(fmt.Sprintf("could not create module page file: %v", err))
		}

		// Render template into page file
		moduleTemplate.Execute(file, moduleTemplateData{
			GoImport: fmt.Sprintf("%s/%s git https://github.com/%s/%s", domain, module, githubOrg, module),
			GoSource: fmt.Sprintf(
				`%s/%s _ https://github.com/%s/%s/tree/main{/dir} https://github.com/%s/%s/blob/main{/dir}/{file}#L{line}`,
				domain,
				module,
				githubOrg,
				module,
				githubOrg,
				module,
			),
			Redirect: fmt.Sprintf("https://godoc.org/%s/%s", domain, module),
		})
	}

	{
		// Create index file
		indexFilePath := filepath.Join(outDir, "index.html")
		file, err := os.Create(indexFilePath)
		defer file.Close()
		if err != nil {
			panic(fmt.Sprintf("could not create index file: %v", err))
		}

		// Render template into index file
		indexTemplate.Execute(file, indexTemplateData{
			Redirect: fmt.Sprintf("https://github.com/%s", githubOrg),
		})
	}

	{
		// Create redirects file
		redirectsFilePath := filepath.Join(outDir, "_redirects")
		file, err := os.Create(redirectsFilePath)
		defer file.Close()
		if err != nil {
			panic(fmt.Sprintf("could not create redirects file: %v", err))
		}

		redirectsTemplate.Execute(file, redirectsTemplateData{
			Modules: modules,
		})
	}
}
