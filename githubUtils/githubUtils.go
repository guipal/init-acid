package githubUtils


import (
"bufio"
"bytes"
"fmt"
"net/http"
"os"
"strings"
"sync"

"github.com/jorgechato/acictl/utils/httpclient"
)

var (
	user = "jorgechato"
	pass = "a3ad3d49975c794d6f021bc1ed1cf360f157797a"
	org  = "adidasMADHackathon"

	messages = make(chan int)
	wg       sync.WaitGroup
)



func main() {
	readFile("repos.txt")
	wg.Wait()
}

func readFile(path string) {
	inFile, _ := os.Open(path)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		wg.Add(1)
		go validate(scanner.Text())
	}
}

func validate(repo string) {
	defer wg.Done()
	url := fmt.Sprintf(
		"https://api.github.com/orgs/%v/repos",
		org,
	)

	c := httpclient.NewClient(
		url,
		"GET",
		"application/json",
		"",
		"",
	)

	c.Run(
		bytes.Buffer{},
		func(res *http.Response, b string) {
			found := strings.Contains(b, repo)

			if !found {
				wg.Add(1)
				go fork(repo)
			} else {
				fmt.Println("repo found")
			}
		},
		true,
		false,
	)
}

func fork(repo string) {
	defer wg.Done()
	url := fmt.Sprintf(
		"https://api.github.com/repos/%v/forks?org=%v",
		repo,
		org,
	)

	b := []byte(`{"organization":"adidasMADHackathon"}`)

	c := httpclient.NewClient(
		url,
		"POST",
		"application/json",
		user,
		pass,
	)

	c.Run(
		*bytes.NewBuffer(b),
		func(res *http.Response, b string) {},
		true,
		false,
	)
}

func CreateRepo(githubUser,githubtoken,repo,org string) {

	url := fmt.Sprintf(
		"https://api.github.com/orgs/%v/repos",
		org,


	)

	b := []byte("{ \"name\": \""+repo+"\",\"description\": \"Base config repository\", \"homepage\": \"https://github.com\", \"private\": false,\"has_issues\": false, \"has_projects\": false, \"has_wiki\": false }")

	c := httpclient.NewClient(
		url,
		"POST",
		"application/json",
		githubUser,
		githubtoken,
	)

	c.Run(
		*bytes.NewBuffer(b),
		func(res *http.Response, b string) {},
		true,
		false,
	)
}

func CreateRepoKey(githubUser,githubtoken,repo,org,keyTitle,key string) {

	url := fmt.Sprintf(
		"https://api.github.com/repos/%v/%v/keys",
		org,
		repo,
	)

	b := []byte("{ \"title\": \""+keyTitle+"\",\"key\": \""+key+"\",\"read_ony\": false}")

	c := httpclient.NewClient(
		url,
		"POST",
		"application/json",
		githubUser,
		githubtoken,
	)

	c.Run(
		*bytes.NewBuffer(b),
		func(res *http.Response, b string) {},
		true,
		false,
	)
}

func GetOrganization(githubUser,githubtoken,org string) string{

	url := fmt.Sprintf(
		"https://api.github.com/orgs/%v",
		org,
	)


	c := httpclient.NewClient(
		url,
		"GET",
		"application/json",
		githubUser,
		githubtoken,
	)
	var body string

	c.Run(
		bytes.Buffer{},
		func(res *http.Response, b string) {
			body=b
		},
		true,
		false,
	)
	return body
}



func GetRepoKeys(githubUser,githubtoken,repo,org string) string{

	url := fmt.Sprintf(
		"https://api.github.com/repos/%v/%v/keys",
		org,
		repo,
	)


	c := httpclient.NewClient(
		url,
		"GET",
		"application/json",
		githubUser,
		githubtoken,
	)
	var body string

	c.Run(
		bytes.Buffer{},
		func(res *http.Response, b string) {
			body=b
		},
		true,
		false,
	)
	return body
}

