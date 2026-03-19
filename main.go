package main

import (
	"context"
	"fmt"
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
		fmt.Println("coulndt handle contest input", err)
		return
	}

	fmt.Println(">Task")
	fmt.Println(inputs)
	fmt.Println(outputs)
	fmt.Println("End Task<")
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
			problems= append(problems, problem)
		}
	}
	problems = unique(problems)
	for _, problem := range problems {
		handleProblem(ctx, problem)
	}
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
			chromedp.WaitReady(".contests-table.group-contests-container"),
			chromedp.Evaluate(`
			Array.from(document.querySelectorAll("a"))
			.map(a => a.href)
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
				handleContest(ctx, url)
			})
		}
	}
	wg.Wait()
}
