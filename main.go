package main

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/chromedp/chromedp"

	//	"sync"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"))

	allocCtx, alocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer alocCancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	url := "https://codeforces.com/group/lwQWQuob0B/contests"

	var contestURL []string
	var contestNames []string
	var contestRemainingTime []string
	err := chromedp.Run(ctx, chromedp.Navigate(url),
				chromedp.WaitReady("tr[data-groupcontestid]"),
				chromedp.Evaluate(`
				Array.from(document.querySelectorAll('tr[data-groupcontestid]'))
				.map(contest => contest.children[0].children[1].href)
					`, &contestURL),

				chromedp.Evaluate(`
				Array.from(document.querySelectorAll('tr[data-groupcontestid]')).
				map(contest => contest.children[0].firstChild.data)	
				`, &contestNames),
				chromedp.Evaluate(`
				Array.from(document.querySelectorAll('tr[data-groupcontestid]'))
				.map(contest => contest.children[3]?.children[2])
				.map(standings => standings?.children[0]?.children[0]?.title)	
				`, &contestRemainingTime))

	if err != nil {
		fmt.Printf("couldn't visit given url: %v and extract contests %v", url, err)
		return
	}
	n := len(contestNames)
	for i := range n {
		fmt.Printf("%v) %v ", i + 1, strings.TrimSpace(contestNames[i]))
		if contestRemainingTime[i] != "" {
			fmt.Printf("(remaining time: %v)", contestRemainingTime[i])
		}
		fmt.Println()
	}
	
	fmt.Printf("Specify contest number to scrape(1 - %v):\n", n)
	var pick int
	_, err = fmt.Scan(&pick)
	pick--
	if err != nil {
		fmt.Printf("Plese provide valid number. err: %v", err)
		return
	}
	handleContest(ctx, contestURL[pick])
}

func handleContest(parent context.Context, contest string) {
	ctx, cancel := chromedp.NewContext(parent)
	defer cancel()
	problemPattern := `/problem/\w+/?$`
	var tasks []string
	err := chromedp.Run(ctx, chromedp.Navigate(contest),
		chromedp.WaitReady(".problems"),
		chromedp.Evaluate(`
						Array.from(document.querySelectorAll("a"))
						.map(a => a.href)
						`, &tasks))
	if err != nil {
		fmt.Println("coulndt handle contest", err)
		return
	}

	var problems []string
	for _, problem := range tasks {
		match, _ := regexp.MatchString(problemPattern, problem)
		if match {
			problems = append(problems, problem)
		}
	}
	problems = makeUnique(problems)
	var wg sync.WaitGroup
	for i, problem := range problems {
		wg.Go(func() {
			handleProblem(fmt.Sprintf("problem%d", i + 1), ctx, problem)
		})
	}
	wg.Wait()
}

func handleProblem(name string, parent context.Context, problem string) {
	ctx, cancel := chromedp.NewContext(parent)
	defer cancel()

	var inputs []string
	var outputs []string

	err := chromedp.Run(ctx, chromedp.Navigate(problem),
		chromedp.WaitReady("div.input"),

		chromedp.Evaluate(`
				Array.from(document.querySelectorAll("div.input"))
				.map((title) => title.children[1].innerText)
				`, &inputs),

		chromedp.WaitReady("div.output"),
		chromedp.Evaluate(`
				Array.from(document.querySelectorAll("div.output"))
				.map((title) => title.children[1].innerText)
				`, &outputs))

	if err != nil {
		fmt.Println("coulnd't handle contest IO", err)
		return
	}
	writeData(name, inputs, outputs)
}

func writeData(name string, input []string, output []string) {
	folderPath := filepath.Join("tests", name)
	err := os.MkdirAll(folderPath, 0755)
	if err != nil {
		fmt.Println("couldn't create a tests directory")
		return
	}

	for i := range len(input) {
		path := filepath.Join(folderPath, fmt.Sprintf("input%d", i + 1))
		f, err := os.Create(path)
		if err != nil {
			fmt.Printf("couldn't create file %v \n", path)
		}
		f.WriteString(input[i])
		f.Close()
		path = filepath.Join(folderPath, fmt.Sprintf("output%d", i + 1))
		f, err = os.Create(path)
		if err != nil {
			fmt.Printf("couldn't create file %v \n", path)
		}
		f.WriteString(output[i])
		f.Close()
	}
}

func makeUnique(arr []string) []string {
	var unique []string
	for i := 0; i < len(arr); i += 2 {
		unique = append(unique, arr[i])
	}
	return unique
}

func folderExists(path string) bool {
	status, err := os.Stat(path)
	if err != nil {
		return false
	}
	return status.IsDir()
}
