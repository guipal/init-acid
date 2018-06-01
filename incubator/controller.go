package incubator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

//Get all Jenkins plugins installed with the current version
func GetPlugins(jenkins Jenkins,version bool, callback func([]byte, string)) {
	url := fmt.Sprintf("%v/pluginManager/api/json?depth=1&tree=plugins[shortName,version]",
		jenkins.Uri,
	)

	client(
		ClientPrt{
			Url:         url,
			Method:      "GET",
			ContentType: "application/json",
			User:        jenkins.User,
			Password:    jenkins.Password,
		},
		func(res *http.Response, b string) {
			response := Response{}

			json.Unmarshal([]byte(b), &response)
			var listPlugins []byte

			for _, plugin := range response.Plugins {
				parsePlugin := []byte("")
				if version {
					parsePlugin = []byte(fmt.Sprintf("%s:%s\n", plugin.ShortName, plugin.Version))
				} else {
					parsePlugin = []byte(fmt.Sprintf("%s\n", plugin.ShortName))
				}
				listPlugins = append(listPlugins, parsePlugin...)
			}

			callback(listPlugins, "plugins.txt")
		},
		func() {
			log.Fatal(FatalBitbucketMsg)
		},
	)
}

//BackupPlugins post the plugins list from Jenkins into bitbucket repository
func BackupPlugins(version, tag, force bool) {
	GetPlugins(Jenkins{},version, func(content []byte, filename string) {
		url := fmt.Sprintf("%v/api/1.0/projects/%v/repos/%v/browse/%s",
			bitbucket.Api,
			bitbucket.Project,
			bitbucket.Repo,
			filename,
		)

		form := Form{
			Filename: filename,
			Content:  content,
			Message:  fmt.Sprintf("Update %s", filename),
			Branch:   bitbucket.Branch,
		}
		commitId := getCommitId(filename)
		if id := getCommitId(""); id == commitId {
			if commitId != "" {
				form.SourceCommitId = id
			}
		}

		client(
			ClientPrt{
				Url:         url,
				Method:      "PUT",
				Body:        form,
				ContentType: "multipart/form-data",
				User:        bitbucket.User,
				Password:    bitbucket.Password,
			},
			func(res *http.Response, b string) {
				if tag {
					response := Response{}

					json.Unmarshal([]byte(b), &response)
					AddTag(response.DisplayId, force)
				}
			},
			func() {
				log.Fatal(FatalBitbucketMsg)
			},
		)
	})
}

// AddTag add a tag in a specific commit. force if tag already exists,
// startPoint = DisplayId for the commit
func AddTag(startPoint string, force bool) {
	url := fmt.Sprintf("%v/git/1.0/projects/%v/repos/%v/tags",
		bitbucket.Api,
		bitbucket.Project,
		bitbucket.Repo,
	)

	body := Body{
		Name:       bitbucket.Tag,
		StartPoint: startPoint,
		Message:    fmt.Sprintf("ADD Tag %v", bitbucket.Tag),
		Type:       "ANNOTATED",
		Force:      force,
	}

	client(
		ClientPrt{
			Url:         url,
			Method:      "POST",
			Body:        body,
			ContentType: "application/json",
			User:        bitbucket.User,
			Password:    bitbucket.Password,
		},
		func(res *http.Response, b string) {},
		func() {
			log.Fatal(FatalBitbucketMsg)
		},
	)
}

//GetTag get bitbucket tag by name
func GetTag(force bool) {
	url := fmt.Sprintf("%v/api/1.0/projects/%v/repos/%v/tags/%s",
		bitbucket.Api,
		bitbucket.Project,
		bitbucket.Repo,
		bitbucket.Tag,
	)

	client(
		ClientPrt{
			Url:         url,
			Method:      "GET",
			ContentType: "application/json",
			User:        bitbucket.User,
			Password:    bitbucket.Password,
		},
		func(res *http.Response, b string) {
			if !force {
				log.Fatal(ForceTagMsg)
			}
		},
		func() {},
	)
}

//GetLastCommit get the DisplayId from the last commit
func GetLastCommitId() string {
	url := fmt.Sprintf("%v/api/1.0/projects/%v/repos/%v/commits?limit=1&until=%v",
		bitbucket.Api,
		bitbucket.Project,
		bitbucket.Repo,
		bitbucket.Branch,
	)

	id := make(chan string, 1)

	client(
		ClientPrt{
			Url:         url,
			Method:      "GET",
			ContentType: "application/json",
			User:        bitbucket.User,
			Password:    bitbucket.Password,
		},
		func(res *http.Response, resBody string) {
			response := Response{}

			json.Unmarshal([]byte(resBody), &response)
			if len(response.Values) > 0 {
				id <- response.Values[0].DisplayId
			} else {
				id <- ""
			}
		},
		func() {
			log.Fatal(FatalBitbucketMsg)
		},
	)

	commit := <-id
	return commit
}

//Get the last commit ID if the file exist
func getCommitId(filename string) string {
	url := fmt.Sprintf("%v/api/1.0/projects/%v/repos/%v/commits?limit=1&until=%v&path=%s",
		bitbucket.Api,
		bitbucket.Project,
		bitbucket.Repo,
		bitbucket.Branch,
		filename,
	)

	id := make(chan string, 1)

	client(
		ClientPrt{
			Url:         url,
			Method:      "GET",
			ContentType: "application/json",
			User:        bitbucket.User,
			Password:    bitbucket.Password,
		},
		func(res *http.Response, resBody string) {
			response := Response{}

			json.Unmarshal([]byte(resBody), &response)
			if len(response.Values) > 0 {
				id <- response.Values[0].DisplayId
			} else {
				id <- ""
			}
		},
		func() {
			log.Fatal(FatalBitbucketMsg)
		},
	)

	commit := <-id
	return commit
}

//CreateBranch create a new branch in the BaseProject
func CreateBranch() {
	url := fmt.Sprintf("%v/branch-utils/1.0/projects/%v/repos/%v/branches",
		bitbucket.Api,
		bitbucket.BaseProject,
		bitbucket.Repo,
	)

	body := Body{
		Name:       bitbucket.Branch,
		StartPoint: bitbucket.OriginBranch,
	}

	client(
		ClientPrt{
			Url:         url,
			Method:      "POST",
			Body:        body,
			ContentType: "application/json",
			User:        bitbucket.User,
			Password:    bitbucket.Password,
		},
		func(res *http.Response, b string) {
			addGroup(bitbucket.Project)
		},
		func() {
			log.Fatal(FatalBitbucketMsg)
		},
	)
}

//Create a group permission in specific branch
func addGroup(project string) {
	url := fmt.Sprintf("%v/branch-permissions/2.0/projects/%v/repos/%v/restrictions",
		bitbucket.Api,
		project,
		bitbucket.Repo,
	)

	body := Body{
		Matcher: Matcher{
			Id:   fmt.Sprintf("refs/heads/%v", bitbucket.Branch),
			Type: Type{Id: "BRANCH", Name: "Branch"},
		},
		Type:   "read-only",
		Groups: []string{bitbucket.Group},
	}

	client(
		ClientPrt{
			Url:         url,
			Method:      "POST",
			Body:        body,
			ContentType: "application/json",
			User:        bitbucket.User,
			Password:    bitbucket.Password,
		},
		func(res *http.Response, b string) {},
		func() {
			//If there was a problem to create group permissions to a branch, lets clean that branch
			DeleteBranch(bitbucket.BaseProject, false)
			DeleteBranch(bitbucket.Project, false)
		},
	)
}

//DeleteBranch delete a branch in bitbucket
func DeleteBranch(project string, clearGroup bool) {
	url := fmt.Sprintf("%v/branch-utils/1.0/projects/%v/repos/%v/branches",
		bitbucket.Api,
		project,
		bitbucket.Repo,
	)

	body := Body{
		Name: bitbucket.Branch,
	}

	client(
		ClientPrt{
			Url:         url,
			Method:      "DELETE",
			Body:        body,
			ContentType: "application/json",
			User:        bitbucket.User,
			Password:    bitbucket.Password,
		},
		func(res *http.Response, b string) {
			if clearGroup {
				getRestrictions()
			}
		},
		func() {
			log.Fatal(FatalBitbucketMsg)
		},
	)
}

//Get group credentials in a branch
func getRestrictions() {
	url := fmt.Sprintf("%v/branch-permissions/2.0/projects/%v/repos/%v/restrictions?matcherType=BRANCH&matcherId=%s",
		bitbucket.Api,
		bitbucket.Project,
		bitbucket.Repo,
		fmt.Sprintf("refs/heads/%v", bitbucket.Branch),
	)

	client(
		ClientPrt{
			Url:         url,
			Method:      "GET",
			ContentType: "application/json",
			User:        bitbucket.User,
			Password:    bitbucket.Password,
		},
		func(res *http.Response, b string) {
			response := Response{}

			json.Unmarshal([]byte(b), &response)
			for _, group := range response.Values {
				clearRestrictions(group.Id)
			}
		},
		func() {
			log.Fatal(FatalBitbucketMsg)
		},
	)
}

//Clear unactive group credentials in a branch
func clearRestrictions(id int) {
	url := fmt.Sprintf("%v/branch-permissions/2.0/projects/%v/repos/%v/restrictions/%d",
		bitbucket.Api,
		bitbucket.Project,
		bitbucket.Repo,
		id,
	)

	client(
		ClientPrt{
			Url:         url,
			Method:      "DELETE",
			ContentType: "application/json",
			User:        bitbucket.User,
			Password:    bitbucket.Password,
		},
		func(res *http.Response, b string) {},
		func() {
			log.Fatal(FatalBitbucketMsg)
		},
	)
}

//DestroyRepo Remove the repository
func DestroyRepo() {
	url := fmt.Sprintf("%v/api/1.0/projects/%v/repos/%v",
		bitbucket.Api,
		bitbucket.Project,
		bitbucket.Repo,
	)

	client(
		ClientPrt{
			Url:         url,
			Method:      "DELETE",
			ContentType: "application/json",
			User:        bitbucket.User,
			Password:    bitbucket.Password,
		},
		func(res *http.Response, b string) {},
		func() {
			log.Fatal(FatalBitbucketMsg)
		},
	)
}
