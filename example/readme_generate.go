package example

import (
	"os"
	"text/template"
	"time"
)

type ReadmeData struct {
	ProjectName      string
	Description      string
	Prerequisites    string
	InstallSteps     string
	UsageExamples    string
	TestInstructions string
	DeployNotes      string
	BuiltWith        string
	ContributingLink string
	Versioning       string
	Authors          string
	License          string
	Acknowledgments  string
}

type ContribData struct {
	ProjectName string
	Date        string
}

func main() {
	readme_data := ReadmeData{
		ProjectName:      "AwesomeCLI",
		Description:      "A cmd‑line tool to turbo‑charge batch image processing",
		Prerequisites:    "- Go 1.21+\n- ImageMagick",
		InstallSteps:     "```bash\ngo install github.com/me/awesomecli@latest\n```",
		UsageExamples:    "```bash\nawesomecli -i imgs/ -o out/ -resize 800x600\n```",
		TestInstructions: "`go test ./...`",
		DeployNotes:      "N/A",
		BuiltWith:        "- Go\n- cobra",
		ContributingLink: "See CONTRIBUTING.md",
		Versioning:       "Semantic Versioning via Git tags",
		Authors:          "- You",
		License:          "MIT License",
		Acknowledgments:  "- Inspired by PurpleBooth’s README template",
	}

	contrib_data := ContribData{
		ProjectName: "AwesomeCLI",
		Date:        time.Now().Format("2006-01-02"),
	}

	generate_doc(readme_data, "readme.tmpl", "README-example.md")
	generate_doc(contrib_data, "contribution.tmpl", "CONTRIBUTING-example.md")
}

func generate_doc(data any, template_file, target_file string) {
	tmpl, err := template.ParseFiles(template_file)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(target_file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		panic(err)
	}
}
