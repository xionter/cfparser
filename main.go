package main

import (
	"context"
	"fmt"
//	"time"
	"regexp"
	"sync"
	"github.com/chromedp/chromedp"
)

func unique(arr []string) []string{
	var unique []string
	for i := 0; i < len(arr); i += 2 {
		unique = append(unique, arr[i])
	}
	return unique
}

func handleProblem(parent context.Context, problem string) {
}

func handleContest(parent context.Context, contest string) {
	ctx, cancel := chromedp.NewContext(parent)
	problemPattern := `/problem/\w+/?$`
	var tasks []string
	err := chromedp.Run(ctx, chromedp.Navigate(contest),
						chromedp.WaitVisible(".problems"),
						chromedp.Evaluate(`
						Array.from(document.querySelectorAll("a")).map(a => a.href)
						`, &tasks))
	cancel()
	if err != nil {
		fmt.Println("coulndt handle contest", err)
		return
	}

	var problems []string
	for _, problem := range tasks {
		match, _ := regexp.MatchString(problemPattern, problem)
		if match {
			problems= append(problems, problem)
		}
	}
	problems = unique(problems)
	
	var wg sync.WaitGroup
	for _, problem := range problems {
		wg.Go(func() {
			fmt.Println(problem)
			handleProblem(ctx, problem)
		})
	}
	wg.Wait()
}

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

	var urls []string
	err := chromedp.Run(ctx, chromedp.Navigate(url),
			chromedp.WaitVisible(".contests-table.group-contests-container"),
			chromedp.Evaluate(`
			Array.from(document.querySelectorAll("a")).map(a => a.href)
			`, &urls))

	if err != nil {
		fmt.Printf("couldn't visit given url: %v and extract contests", url)
		return
	}

	contestPattern := `/contest/\d+/?$`
	var wg sync.WaitGroup
	for _, url := range urls {
		match, _ := regexp.MatchString(contestPattern, url)
		if match {
			wg.Go(func() {
				fmt.Printf("Contest#%v\n", url)
				handleContest(ctx, url)
			})
		}
	}
	wg.Wait()
}
