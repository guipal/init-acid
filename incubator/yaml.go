package incubator

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func (c *YamlConfig) readYaml(path string) {
	cerr := isFile(path)

	check(cerr, func() {
		yamlFile, _ := ioutil.ReadFile(path)
		yaml.Unmarshal(yamlFile, &c)
	})
}

func isFile(filename string) error {
	var e error

	if _, e = os.Stat(filename); !os.IsNotExist(e) {
		return e
	}

	return e
}
