package incubator

type ClientPrt struct {
	Url         string
	Method      string
	ContentType string
	Password    string
	User        string
	Body        interface{}
}

type YamlConfig struct {
	Jenkins   Jenkins   `yaml:"jenkins,omitempty"`
	Bitbucket Bitbucket `yaml:"bitbucket,omitempty"`
}

//Jenkins struct
type Jenkins struct {
	Uri      string `yaml:"uri,omitempty"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
}

//Bitbucket struct
type Bitbucket struct {
	Repo         string `yaml:"repo,omitempty"`
	User         string `yaml:"user,omitempty"`
	Password     string `yaml:"password,omitempty"`
	Group        string `yaml:"group,omitempty"`
	Project      string `yaml:"project,omitempty"`
	BaseProject  string `yaml:"-"`
	Branch       string `yaml:"-"`
	OriginBranch string `yaml:"-"`
	Api          string `yaml:"api,omitempty"`
	Tag          string `yaml:"-"`
}

//HTTP form request struct
type Form struct {
	Filename       string `json:"-"`
	Content        []byte `json:"content,omitempty"`
	Message        string `json:"message,omitempty"`
	Branch         string `json:"branch,omitempty"`
	SourceCommitId string `json:"sourceCommitId,omitempty"`
}

//HTTP body request struct
type Body struct {
	Message    string   `json:"message,omitempty"`
	Name       string   `json:"name,omitempty"`
	StartPoint string   `json:"startPoint,omitempty"`
	Type       string   `json:"type,omitempty"`
	Force      bool     `json:"force,omitempty"`
	Groups     []string `json:"groups,omitempty"`
	Matcher    `json:"matcher,omitempty"`
}

type Matcher struct {
	Id     string `json:"id,omitempty"`
	Active bool   `json:"active,omitempty"`
	Type   `json:"type,omitempty"`
}

type Type struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

//HTTP body response struct
type Response struct {
	Values    []Values `json:"values,omitempty"`
	DisplayId string   `json:"displayId,omitempty"`
	Id        string   `json:"id,omitempty"`
	LastSync  int      `json:"lastSync,omitempty"`
	Plugins   []struct {
		ShortName string `json:"shortName,omitempty"`
		Version   string `json:"version,omitempty"`
	} `json:"plugins,omitempty"`
}

type Values struct {
	DisplayId string `json:"displayId"`
	Id        int    `json:"id"`
}

//Global variables
var (
	jenkins   = Jenkins{}
	bitbucket = Bitbucket{}
)

//Util functions
func (j *Jenkins) isEmpty() bool {
	if j.User == "" || j.Uri == "" || j.Password == "" {
		return true
	}
	return false
}

func (b *Bitbucket) isEmpty() bool {
	if b.User == "" || b.Api == "" || b.Password == "" {
		return true
	}
	return false
}

func (s *Body) isEmpty() bool {
	if s.Name != "" || s.Type != "" {
		return false
	}
	return true
}

func (s *Form) isEmpty() bool {
	if s.Message == "" && s.Branch == "" {
		return true
	}
	return false
}

func (s *Form) isCommit() bool {
	if s.SourceCommitId != "" {
		return true
	}
	return false
}

func NewIncubator(configFile string, models ...interface{}) {
	for _, model := range models {
		switch model.(type) {
		case Bitbucket:
			model.(Bitbucket).create(configFile)
		case Jenkins:
			model.(Jenkins).create(configFile)
		}
	}
}

func (b Bitbucket) create(configFile string) {
	bitbucket = b

	if configFile != "" {
		yamlConfig := YamlConfig{
			Bitbucket: bitbucket,
		}

		yamlConfig.readYaml(configFile)

		if !yamlConfig.Bitbucket.isEmpty() {
			bitbucket = yamlConfig.Bitbucket
		}
	}

	if b.Project != "" {
		bitbucket.Project = b.Project
	}
	if b.Repo != "" {
		bitbucket.Repo = b.Repo
	}
}

func (j Jenkins) create(configFile string) {
	jenkins = j

	if configFile != "" {
		yamlConfig := YamlConfig{
			Jenkins: jenkins,
		}

		yamlConfig.readYaml(configFile)

		if !yamlConfig.Jenkins.isEmpty() {
			jenkins = yamlConfig.Jenkins
		}
	}
}
