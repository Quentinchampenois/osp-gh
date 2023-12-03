package main

import (
	"fmt"
	"github.com/cli/go-gh/v2"
	"log"
	"strings"
	"time"
)

type Repository struct {
	Name  string
	Owner string
	PRs   []PR
}

func (r *Repository) addPR(pr PR) {
	r.PRs = append(r.PRs, pr)
}

func (r *Repository) resume() {
	fmt.Printf("%s \n   Number of PRs : %d\n", r.Name, len(r.PRs))
	for _, pr := range r.PRs {
		if pr.isOpened() {
			dt, _ := time.Parse(time.DateTime, pr.Date)

			fmt.Printf("   #%s ready for review - %s \n", pr.ID, dt.Format("01-02-2006"))
			fmt.Printf("       https://github.com/%s/pull/%s \n", r.Name, pr.ID)
		} else {
			fmt.Printf("   #%s not ready - %s \n", pr.ID, pr.Date)
			parse, _ := time.Parse(pr.Date, "01-12-2020")
			fmt.Printf("%s", parse)
		}
	}
}

type PR struct {
	ID     string
	Title  string
	Branch string
	State  string
	Date   string
}

func (pr *PR) Init(str []string) {
	pr.ID = str[0]
	pr.Title = str[1]
	pr.Branch = str[2]
	pr.State = str[3]
	pr.Date = str[4]
}

func (pr *PR) isOpened() bool {
	if pr.State == "OPEN" {
		return true
	}
	return false
}

func splitOutput(output string) []string {
	return strings.Split(output, "\n")
}

func splitLine(line string) []string {
	return strings.Split(line, "\t")
}

func main() {
	owner := "OpenSourcePolitics"
	ghReposCmd, _, err := gh.Exec("repo", "list", owner, "--limit", "10")
	if err != nil {
		log.Fatal(err)
	}
	repos := splitOutput(ghReposCmd.String())
	var repositories []Repository

	for _, repo := range repos {
		if repo == "" {
			continue
		}

		currentRepository := Repository{Name: splitLine(repo)[0], Owner: owner}

		prList, _, err := gh.Exec("pr", "list", "--repo", currentRepository.Name, "--limit", "15")
		if err != nil {
			log.Fatal(err)
		}

		str := prList.String()
		prs := splitOutput(str)

		for _, pr := range prs {
			if pr == "" {
				continue
			}
			pullRequest := PR{}
			pullRequest.Init(splitLine(pr))
			currentRepository.addPR(pullRequest)
		}

		repositories = append(repositories, currentRepository)
	}

	for _, repo := range repositories {
		repo.resume()
	}
}
