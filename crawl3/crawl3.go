package crawl3

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
	} else {
		return og
	}
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

func getHtml(link, og string, ogLinkChan, nogLinkChan, doneChan chan string) {
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
	og := correctLink("http://www.therileyjohnson.com.s3-website.us-east-2.amazonaws.com/")
	//emails := make(map[string]bool)
	iLinks := map[string]bool{og: false}
	oLinks := make(map[string]bool)
	finLinks := map[string]bool{og: false}
	ogChl := make(chan string)
	nogChl := make(chan string)
	doneChl := make(chan string)
	//emailRegi := regexp.MustCompile(`[\w\d].*@.*\.\w`)
	//linkRegi := regexp.MustCompile(`href=['"]?([^'" >]+)`)
	var link, currentUrl, lastUrl string
	for {
		lastUrl = currentUrl
		if currentUrl = fetchFirst(iLinks); currentUrl == "" && finLinks[lastUrl] {
			break
		}
		iLinks[currentUrl] = true
		getHtml(currentUrl, og, ogChl, nogChl, doneChl)
		for !finLinks[currentUrl] {
			select {
			case link = <-ogChl:
				_, bexists := iLinks[link]
				if !bexists {
					iLinks[link] = false
					finLinks[link] = false
				}
			case link = <-nogChl:
				oLinks[link] = true
			case link = <-doneChl:
				finLinks[link] = true
			}
		}
	}
	for k, _ := range oLinks {
		fmt.Println(k)
	}
	for k, v := range iLinks {
		fmt.Println(v, k)
	}
	fmt.Println("done")
}
