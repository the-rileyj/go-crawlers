package crawl1

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

func getLinks(html, og string, r *regexp.Regexp, ogLinkChan, nogLinkChan chan string, controlChan chan bool){
	var buffer bytes.Buffer
	for _, links := range r.FindAllStringSubmatch(html, -1){
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
	controlChan <- false
}

func main(){
	og := correctLink("http://dsu.edu/")
	resp, err := http.Get(og)
	if err != nil {		
		fmt.Println(err)
	} else {
		//emails := make(map[string]bool)
		iLinks := map[string]bool{ og: true }
		oLinks := make(map[string]bool)
		responseData, err := ioutil.ReadAll(resp.Body)
		//emailRegi := regexp.MustCompile(`[\w\d].*@.*\.\w`)
		linkRegi := regexp.MustCompile(`href=['"]?([^'" >]+)`)
		if err != nil {
    			fmt.Println(err)
		}
		conChl := make(chan bool)
		ogChl := make(chan string)
		nogChl := make(chan string)
		var link string
		go getLinks(string(responseData), og, linkRegi, ogChl, nogChl, conChl)
		//for link := range ogChl{
			//fmt.Println(link)	
		//}	
		brake := true
		for brake{
			select{
				case link = <- ogChl:
					_, berr := iLinks[link]
					if !berr{
						iLinks[link] = false
					}
				case link = <- nogChl:
					_, berr := oLinks[link]
					if !berr{
						oLinks[link] = true
					}
				case brake = <- conChl:
			}
		}	
		for k, _ := range oLinks{
			fmt.Println(k)
		}	
		for k, _ := range iLinks{
			fmt.Println(k)
		}
		fmt.Println("done")
	}
}