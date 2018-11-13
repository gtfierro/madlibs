package main

import (
	//"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
)

type MadlibTemplate struct {
	Template *template.Template
	Title    string
	Prompts  []string
}

type Madlib struct {
	Template *template.Template
	Title    string
	Prompts  []string
	Answers  []string
}

func NewMadlibTemplate(title string, segments []string, prompts []string) (MadlibTemplate, error) {

	var b strings.Builder
	for idx, segment := range segments {
		b.WriteString(segment)
		if idx < len(prompts) {
			b.WriteString(fmt.Sprintf("{{ index .Answers %d }}", idx))
		}
	}

	templateString := b.String()

	madlibtempl, err := template.New(title).Parse(templateString)
	if err != nil {
		return MadlibTemplate{}, err
	}

	return MadlibTemplate{
		Title:    title,
		Template: madlibtempl,
		Prompts:  prompts,
	}, nil
}

func (tmpl MadlibTemplate) NewMadlib() (*Madlib, error) {
	newt, err := tmpl.Template.Clone()
	if err != nil {
		return nil, err
	}
	return &Madlib{
		Template: newt,
		Title:    tmpl.Title,
		Prompts:  tmpl.Prompts,
	}, nil
}

func (madlib *Madlib) HasNextPrompt() bool {
	return len(madlib.Answers) != len(madlib.Prompts)
}

func (madlib *Madlib) NextPrompt() string {
	return madlib.Prompts[len(madlib.Answers)]
}

func (madlib *Madlib) AddAnswer(s string) {
	madlib.Answers = append(madlib.Answers, strings.TrimSpace(s))
}

func (madlib *Madlib) Finish() string {
	var b strings.Builder
	err := madlib.Template.Execute(&b, madlib)
	if err != nil {
		return err.Error()
	}
	return b.String()
}

func main() {
	f, err := os.Open("madlibs.json")
	if err != nil {
		log.Fatal(err)
	}

	dec := json.NewDecoder(f)
	var madlibs []struct {
		Title    string
		Prompts  []string
		Segments []string
	}
	err = dec.Decode(&madlibs)
	if err != nil {
		log.Fatal(err)
	}

	var templates []MadlibTemplate
	for _, t := range madlibs {
		ml, err := NewMadlibTemplate(t.Title, t.Segments, t.Prompts)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(t.Title, len(t.Segments), len(t.Prompts))
		templates = append(templates, ml)
	}

	srv := NewMadlibServer(":8999", templates)
	srv.ServeMadlibs()
}
