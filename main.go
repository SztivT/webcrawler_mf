package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"os"
	"strconv"
)

const host = "https://breakit.se"

var visited []string

func main() {
	depth, concurrentRequests := cliArgs()
	requestChannel := make(chan string, concurrentRequests)
	crawl("/", depth, requestChannel)
}

func crawl(query string, depth int, requestChannel chan string) {
	if depth == -1 {
		return
	}
	request, err := http.Get(host + query)
	if request == nil {
		return
	}
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(request.Body)
	if request.StatusCode != 200 {
		fmt.Println(request.Status)
		return
	}
	document, err := goquery.NewDocumentFromReader(request.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	document.Find("body").Each(func(i int, selection *goquery.Selection) {
		selection.Find("a").Each(func(i int, selection *goquery.Selection) {
			href, _ := selection.Attr("href")
			if isVisited(href) {
				return
			}
			requestChannel <- href
			go crawl(href, depth-1, requestChannel)
			displayContent(<-requestChannel)
		})
	})
}

func isVisited(href string) bool {
	for _, val := range visited {
		if val == href {
			return true
		}
	}
	visited = append(visited, href)
	return false
}

func displayContent(query string) {
	if len(query) < 9 {
		return
	}
	if query[0:8] != "/artikel" {
		return
	}
	request, err := http.Get(host + query)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(request.Body)
	if request.StatusCode != 200 {
		fmt.Println(request.Status)
		return
	}
	document, err := goquery.NewDocumentFromReader(request.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("\nArticle details:\n")
	fmt.Printf("URL: %s\n", host+query)
	dateTime := document.Find("time").Text()
	fmt.Printf("Publising Date: %s\n", dateTime)
	mainTitle := document.Find("h1").Text()
	fmt.Printf("Main title: %s\n", mainTitle)
	subTitle := document.Find("h4").Text()
	fmt.Printf("Subtitle: %s\n", subTitle)
	firstParagraph := document.Find("p").First().Text()
	fmt.Printf("First paragraph: %s\n", firstParagraph)
}

func cliArgs() (int, int) {
	var err error
	depth := 0
	requests := 0
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "-d" {
			depth, err = strconv.Atoi(args[i+1])
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		if args[i] == "-p" {
			requests, err = strconv.Atoi(args[i+1])
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
	return depth, requests
}
