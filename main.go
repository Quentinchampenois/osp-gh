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

func (r *Repository) resumeInLines() {
	fmt.Printf("%s (PRs: %d) \n", r.Name, len(r.PRs))
	for _, pr := range r.PRs {
		parsedDate, _ := time.Parse(time.RFC3339, pr.Date)
		creationDateAgo := timeAgo(parsedDate)
		fmt.Printf("* #%s %s - %s %s \n", pr.ID, pr.Title, pr.State, creationDateAgo)
	}
}

func (r *Repository) resume() {
	fmt.Printf("%s \n   Number of PRs : %d\n", r.Name, len(r.PRs))
	for _, pr := range r.PRs {
		parsedDate, _ := time.Parse(time.RFC3339, pr.Date)
		creationDateAgo := timeAgo(parsedDate)
		if pr.isOpened() {
			fmt.Printf("   #%s Ready - %s \n", pr.ID, creationDateAgo)
			fmt.Printf("       https://github.com/%s/pull/%s \n", r.Name, pr.ID)
		} else {
			fmt.Printf("   #%s not ready - %s \n", pr.ID, creationDateAgo)
		}

		staleDate := time.Now().AddDate(0, 0, -20)
		if parsedDate.Before(staleDate) {
			fmt.Printf("   #%s staled - %s \n", pr.ID, creationDateAgo)
		}
	}
}

// Check if repository has opened PRs
func (r *Repository) hasOpenedPRs() bool {
	for _, pr := range r.PRs {
		if pr.isOpened() {
			return true
		}
	}
	return false
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

func (pr *PR) isValidCI(repo string) bool {
	ghReposCmd, _, err := gh.Exec("pr", "--repo", repo, "checks", pr.ID, "--fail-fast", "--watch")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ghReposCmd.String())
	str := ghReposCmd.String()
	prs := splitOutput(str)
	fmt.Println(prs[len(prs)-3])
	return true
}

func splitOutput(output string) []string {
	return strings.Split(output, "\n")
}

func splitLine(line string) []string {
	return strings.Split(line, "\t")
}

func timeAgo(then time.Time) string {
	now := time.Now()
	diff := now.Sub(then)

	switch {
	case diff < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	case diff < time.Hour*24:
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	default:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	}
}

func main() {
	owner := "OpenSourcePolitics"
	ghReposCmd, _, err := gh.Exec("repo", "list", owner, "--limit", "80")
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
		if !repo.hasOpenedPRs() {
			continue
		}
		repo.resumeInLines()
	}
}
