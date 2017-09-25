package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

func correctLink(og string) string {
	if []rune(og)[len(og)-1] == '/' {
		return string([]rune(og)[:len(og)-1])
	}
	return og
}

func isInternalOG(link string) bool {
	if string([]rune(link)[:4]) != "http" {
		return true
	}
	return false
}

func isPartOG(link, og string) bool {
	nog := []rune(link)
	for i, ogChar := range []rune(og) {
		if nog[i] != ogChar {
			return false
		}
	}
	return true
}

func getLinks(html, link, og string, ogLinkChan, nogLinkChan, doneChan chan string) {
	r := regexp.MustCompile(`href=['"]?([^'" >]+)['"]`)
	var buffer bytes.Buffer
	for _, links := range r.FindAllStringSubmatch(html, -1) {
		if []rune(links[1])[0] == '/' {
			buffer.WriteString(og)
			buffer.WriteString(links[1])
			ogLinkChan <- buffer.String()
		} else if isInternalOG(links[1]) {
			buffer.WriteString(og)
			buffer.WriteString("/")
			buffer.WriteString(links[1])
			ogLinkChan <- buffer.String()
		} else if isPartOG(links[1], og) {
			buffer.WriteString(links[1])
			ogLinkChan <- buffer.String()
		} else {
			buffer.WriteString(links[1])
			nogLinkChan <- buffer.String()
		}
		buffer.Reset()
	}
	doneChan <- link
}

func getHTML(link, og string, ogLinkChan, nogLinkChan, doneChan chan string) {
	resp, err := http.Get(link)
	if err != nil {
		fmt.Println(err)
	} else {
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		go getLinks(string(responseData), link, og, ogLinkChan, nogLinkChan, doneChan)
	}
}

func fetchFirst(i map[string]bool) string {
	for k, v := range i {
		if !v {
			return k
		}
	}
	return ""
}

func main() {
	og := correctLink("http://www.dsu.edu/") //Assuring that the original link is in the format required for later comparison
	//og := correctLink("http://www.revengeofficial.com/")
	totalCrawled, currentLinks, speedLimit := 0, 0, 20 //Maximum number of http requests happening at any given time
	//emails := make(map[string]bool)
	//Declaration and initialization of the following maps:
	//TODO
	finLinks := map[string]bool{og: false}
	//Declaration and initialization of the following channels:
	//TODO
	doneChl, nogChl, ogChl := make(chan string), make(chan string), make(chan string)
	//Declaration and initialization of the following slices:
	//TODO
	toCrawl, oLinks := []string{og}, []string{}
	//emailRegi := regexp.MustCompile(`[\w\d].*@.*\.\w`)
	//linkRegi := regexp.MustCompile(`href=['"]?([^'" >]+)`)
	var link, currentURL string
	//lastURL
	for {
		if currentLinks <= speedLimit {
			if len(toCrawl) == 0 {
				if finLinks[currentURL] {
					break
				}
			} else {
				currentURL, toCrawl = toCrawl[0], toCrawl[1:]
				go getHTML(currentURL, og, ogChl, nogChl, doneChl)
				finLinks[link] = false
				currentLinks++
				totalCrawled++
			}
		}
		select {
		case link = <-ogChl:
			_, bexists := finLinks[link]
			if !bexists {
				println(bexists)
				toCrawl = append(toCrawl, link)
			}
		case link = <-nogChl:
			oLinks = append(oLinks, link)
		case link = <-doneChl:
			currentLinks--
			finLinks[link] = true
		}
	}
	for _, v := range oLinks {
		fmt.Println(v)
	}
	for k := range finLinks {
		fmt.Println(k)
	}
	fmt.Printf("done, total: %d", totalCrawled)
}
