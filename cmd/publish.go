/*
Copyright 2022 EscherCloud.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/drew-viles/baskio/pkg/constants"
	"github.com/drew-viles/baskio/pkg/file"
	gitRepo "github.com/drew-viles/baskio/pkg/git"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	tpl "text/template"
	"time"
)

// fetchPagesRepo pulls the GitHub pages repo locally for modification.
func fetchPagesRepo(ghUser, ghToken, ghProject, ghBranch string) (string, *git.Repository, error) {
	pagesRepo := fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", ghUser, ghToken, ghUser, ghProject)
	pagesDir := filepath.Join("/tmp", ghProject)

	err := os.MkdirAll(pagesDir, 0755)
	if err != nil {
		return "", nil, err
	}

	pagesBranch := plumbing.ReferenceName(filepath.Join("refs/heads", ghBranch))
	g, err := gitRepo.GitClone(pagesRepo, pagesDir, pagesBranch)
	if err != nil {
		return "", nil, fmt.Errorf("git clone error: %s\n", err)
	}

	return pagesDir, g, nil
}

// copyResultsFileIntoPages copies the results of the recent scan into the relevant
// location for the static site to be able to display them.
func copyResultsFileIntoPages(gitDir string, resultsFile *os.File) error {
	log.Println("copying results file into pages repo")
	resultsDir := filepath.Join(gitDir, "docs", "results")
	scanDate := fmt.Sprintf("cve-%d-%d-%d--%d-%d-%d", time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(), time.Now().Minute(), time.Now().Second())
	cveFile := strings.Join([]string{scanDate, "json"}, ".")

	err := os.MkdirAll(resultsDir, 0755)
	if err != nil {
		return err
	}

	_, err = file.CopyFile(resultsFile.Name(), filepath.Join(resultsDir, cveFile))
	if err != nil {
		return err
	}

	return nil
}

// fetchExistingReports runs to collect all reports from the results directory so that they can be parsed for later usage.
func fetchExistingReports(gitDir string) ([]string, error) {
	log.Println("collating existing reports")
	var reportPaths []string

	resultsDir := filepath.Join(gitDir, "docs", "results")

	dirs, err := os.ReadDir(resultsDir)
	if err != nil {
		return nil, err
	}

	for _, v := range dirs {
		if !v.IsDir() {
			reportPaths = append(reportPaths, filepath.Join(resultsDir, v.Name()))
		}
	}

	return reportPaths, nil
}

// parseReports turns all json files into structs to be used in templating for the static site.
func parseReports(reports []string) (map[int]constants.Year, error) {
	log.Println("parsing reports")
	allReports := make(map[int]constants.Year)

	for _, v := range reports {
		var r constants.ReportData
		file, err := os.ReadFile(v)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(file, &r)
		if err != nil {
			return nil, err
		}

		stripDirPrefix := strings.Split(v, "/")
		reportName := stripDirPrefix[len(stripDirPrefix)-1:][0]
		fileName := strings.Split(reportName, "-")

		year, err := strconv.Atoi(fileName[1])
		if err != nil {
			return nil, err
		}
		month, err := strconv.Atoi(fileName[2])
		if err != nil {
			return nil, err
		}
		monthName := time.Month(month).String()

		if _, ok := allReports[year]; !ok {
			allReports[year] = constants.Year{}
		}

		if _, ok := allReports[year].Months[monthName]; !ok {
			y := allReports[year]
			y.Months = make(map[string]constants.Month)
			y.Months[monthName] = constants.Month{}
			allReports[year] = y
		}

		if allReports[year].Months[monthName].Reports == nil {
			m := allReports[year].Months[monthName]
			m.Reports = make(map[string]constants.ReportData)
			m.Reports[reportName] = r
			allReports[year].Months[monthName] = m
		} else {
			allReports[year].Months[monthName].Reports[reportName] = r
		}
	}
	return allReports, nil
}

// buildPages will generate all the pages required for GitHub Pages.
func buildPages(projectDir string, allReports map[int]constants.Year) error {
	log.Println("building pages")
	baseDir := strings.Join([]string{projectDir, "docs"}, "/")
	err := generateHTMLTemplate(baseDir, allReports)
	if err != nil {
		return err
	}

	err = generateJSTemplates(baseDir, allReports)
	if err != nil {
		return err
	}

	return nil
}

// generateHTMLTemplate creates the index.html page for the frontend website which displays the CVE data.
func generateHTMLTemplate(baseDir string, allReports map[int]constants.Year) error {
	log.Println("generating index.html")
	// HTML template
	htmlTarget := strings.Join([]string{baseDir, "index.html"}, "/")
	htmlTmpl := "templates/index.html.gotmpl"
	htmlFile, err := os.Create(htmlTarget)
	if err != nil {
		return err
	}
	t := template.Must(template.New("index.html.gotmpl").Funcs(template.FuncMap{
		"inc": func(x int) int {
			return x + 1
		},
	}).ParseFiles(htmlTmpl))
	if err != nil {
		return err
	}

	err = t.ExecuteTemplate(htmlFile, "index.html.gotmpl", allReports)
	if err != nil {
		return err
	}
	return nil
}

// generateJSTemplates creates all the Javscript files for the frontend website.
func generateJSTemplates(baseDir string, allReports map[int]constants.Year) error {
	jsDir := filepath.Join(baseDir, "js")

	err := os.MkdirAll(jsDir, 0755)
	if err != nil {
		return err
	}

	// main.js template
	log.Println("generating main.js")
	mainJSTarget := filepath.Join(jsDir, "main.js")
	mainJSTmpl := "templates/js/main.js.gotmpl"
	mainFile, err := os.Create(mainJSTarget)
	if err != nil {
		return err
	}

	m := tpl.Must(tpl.New("main.js.gotmpl").Funcs(tpl.FuncMap{
		"inc": func(x int) int {
			return x + 1
		},
	}).ParseFiles(mainJSTmpl))
	if err != nil {
		return err
	}

	err = m.ExecuteTemplate(mainFile, "main.js.gotmpl", allReports)
	if err != nil {
		return err
	}

	// class.js template
	log.Println("generating class.js")
	classJSTarget := filepath.Join(jsDir, "class.js")
	classJSTmpl := "templates/js/class.js.gotmpl"
	classFile, err := os.Create(classJSTarget)
	if err != nil {
		return err
	}

	c := tpl.Must(tpl.New("class.js.gotmpl").Funcs(tpl.FuncMap{
		"inc": func(x int) int {
			return x + 1
		},
	}).ParseFiles(classJSTmpl))
	if err != nil {
		return err
	}

	err = c.ExecuteTemplate(classFile, "class.js.gotmpl", allReports)
	if err != nil {
		return err
	}

	// reports.js template
	log.Println("generating reports.js")
	reportJSTarget := filepath.Join(jsDir, "reports.js")
	reportJSTmpl := "templates/js/reports.js.gotmpl"
	reportFile, err := os.Create(reportJSTarget)
	if err != nil {
		return err
	}

	r := tpl.Must(tpl.New("reports.js.gotmpl").Funcs(tpl.FuncMap{
		"inc": func(x int) int {
			return x + 1
		},
	}).ParseFiles(reportJSTmpl))
	if err != nil {
		return err
	}

	reportNames := []string{}
	for _, re := range allReports {
		for _, mo := range re.Months {
			for k := range mo.Reports {
				reportNames = append(reportNames, k)
			}
		}
	}
	err = r.ExecuteTemplate(reportFile, "reports.js.gotmpl", reportNames)
	if err != nil {
		return err
	}

	return nil
}

// publishPages pushes the generated javascript, html and results file to GitHub pages.
func publishPages(repository *git.Repository, gitPagesPath string) error {
	log.Println("publishing to GitHub pages")

	w, err := repository.Worktree()
	if err != nil {
		return fmt.Errorf("working tree error: %s", err)
	}

	_, err = w.Add("docs")
	if err != nil {
		return fmt.Errorf("adding files error: %s", err)
	}

	auth := &object.Signature{
		Name:  "Openstack Kube Images",
		Email: "openstack-kube-images@github",
		When:  time.Now(),
	}

	commitOptions := &git.CommitOptions{
		All:       false,
		Author:    auth,
		Committer: auth,
	}

	_, err = w.Commit("patch: Adding latest results and updating pages", commitOptions)
	if err != nil {
		return fmt.Errorf("commit error: %s", err)
	}

	err = repository.Push(&git.PushOptions{
		RemoteName: "origin",
	})
	if err != nil {
		return fmt.Errorf("push error %s", err)
	}

	return nil
}

// pagesCleanup just removes any leftover files/repo so that when running locally the binary doesn't hit a conflict.
// This ensures that on multiple runs it always ahas the latest code base for the GitHub pages repo.
func pagesCleanup(pagesDir string) {
	err := os.RemoveAll(pagesDir)
	if err != nil {
		log.Fatalln(err)
	}
}
