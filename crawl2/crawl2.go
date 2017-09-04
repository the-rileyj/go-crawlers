package crawl2

import(
	"net/http"
	"fmt"
	"io/ioutil"
	"regexp"
	"bytes"
)

func correctLink(og string) string{
	if []rune(og)[len(og)-1] == '/'{
		return string([]rune(og)[:len(og)-1])
	} else {
		return og
	}
}

func isPartOG(link, og string) bool{
	nog := []rune(link)
	for i, ogChar := range []rune(og){
		if nog[i] != ogChar{
			return false
		}
	}
	return true
}

func getLinks(link, og string, ogLinkChan, nogLinkChan, doneChan chan string){
	resp, err := http.Get(og)
	if err != nil {		
		fmt.Println(err)
	} else {
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil{
			fmt.Println(err)
		}
		r := regexp.MustCompile(`href=['"]?([^'" >]+)`)
		var buffer bytes.Buffer
		for _, links := range r.FindAllStringSubmatch(string(responseData), -1){
			if []rune(links[1])[0] == '/'{
				buffer.WriteString(og)
				buffer.WriteString(links[1])
				ogLinkChan <- buffer.String()
			} else if isPartOG(links[1], og){
				buffer.WriteString(links[1])
				ogLinkChan <- buffer.String()
			} else {
				buffer.WriteString(links[1])
				nogLinkChan <- buffer.String()
			}
			buffer.Reset()
		}
	}
	doneChan <- link
}

func notDone(i, f map[string]bool)(url string, nfin bool){
	for k, v := range i{
		if v && !f[k] {
			nfin = true	
		} else if !v {
			return k, true
		}
	}
	return
	//implicit: return url, nfin
}

func main(){
	og := correctLink("http://www.therileyjohnson.com.s3-website.us-east-2.amazonaws.com/")
	//emails := make(map[string]bool)
	iLinks := map[string]bool{ og: false }
	oLinks := make(map[string]bool)
	finLinks := map[string]bool{ og: false }
	//emailRegi := regexp.MustCompile(`[\w\d].*@.*\.\w`)
	//linkRegi := regexp.MustCompile(`href=['"]?([^'" >]+)`)
	ogChl := make(chan string)
	nogChl := make(chan string)
	doneChl := make(chan string)
	var link string
	inc := 0
	for{
		url, done := notDone(iLinks, finLinks)
		if done && (url == "" || inc > 8) {
			for inc = inc; inc > 0; inc-- {
				select{
					case link = <- ogChl:
						_, berr := iLinks[link]
						if !berr{
							iLinks[link] = false
							finLinks[link] = false
						} 
					case link = <- nogChl:
						_, berr := oLinks[link]
						if !berr{
							oLinks[link] = true
						}
					case link = <- doneChl:
						finLinks[link] = true
				}
			}
		} else if !done && url == "" {
			break
		} else if done && url != ""{
			iLinks[url] = true
			fmt.Println(url)
			go getLinks(url, og, ogChl, nogChl, doneChl)
			inc++
		}
		fmt.Println(inc)
	}	
	for k, _ := range oLinks{
		fmt.Println(k)
	}	
	for k, _ := range iLinks{
		fmt.Println(k)
	}
	fmt.Println("done")
}