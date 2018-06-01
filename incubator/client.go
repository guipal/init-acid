package incubator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	////Uncomment for debug
	//"io"
	//"os"
)

//Check if something was wrong
func check(e error, callback func()) {
	if e != nil {
		msg := fmt.Sprintf(
			SomethingWentWrong,
			e,
		)
		log.Fatal(msg)
	} else {
		callback()
	}
}

//Bitbucket REST API call function
func client(c ClientPrt, callback func(res *http.Response, b string), errCallback func()) {
	client := http.Client{}
	buf := bytes.Buffer{}

	//Create form if needed otherwise create json body
	switch c.Body.(type) {
	case Body:
		b, _ := json.Marshal(c.Body.(Body))
		buf.WriteString(string(b))
	case Form:
		w := multipart.NewWriter(&buf)

		createForm(c.Body.(Form), w)

		c.ContentType = w.FormDataContentType()
		w.Close()
	}

	fmt.Printf("+ curl -X %s %v <<< %v\n", c.Method, c.Url, buf.String())
	req, err := http.NewRequest(c.Method, c.Url, &buf)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", c.ContentType)
	req.SetBasicAuth(c.User, c.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp.Status)

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	bodyString := string(bodyBytes)

	////Uncomment for debug
	//if _, err := io.Copy(os.Stderr, resp.Body); err != nil {
	//log.Fatal(err)
	//}
	//log.Println()

	if resp.StatusCode < 399 && resp.StatusCode > 100 {
		callback(resp, bodyString)
	} else {
		errCallback()
	}
}

func createForm(f Form, w *multipart.Writer) {
	fw, err := w.CreateFormField("content")
	check(err, func() {
		if _, err = fw.Write(f.Content); err != nil {
			log.Fatal(err)
		}
	})

	if fw, err = w.CreateFormField("message"); err != nil {
		log.Fatal(err)
	}
	if _, err = fw.Write([]byte(f.Message)); err != nil {
		log.Fatal(err)
	}

	if fw, err = w.CreateFormField("branch"); err != nil {
		log.Fatal(err)
	}
	if _, err = fw.Write([]byte(f.Branch)); err != nil {
		log.Fatal(err)
	}

	if f.isCommit() {
		if fw, err = w.CreateFormField("sourceCommitId"); err != nil {
			log.Fatal(err)
		}
		if _, err = fw.Write([]byte(f.SourceCommitId)); err != nil {
			log.Fatal(err)
		}
	}
}
