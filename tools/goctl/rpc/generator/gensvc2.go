package generator

import (
	"fmt"
	"io/ioutil"
	"regexp"
)

func text2Template(path string) (text string, err error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return replaceTags(content)
}
func replaceTags(content []byte) (str string, err error) {
	tags := map[string]string{
		"codeGeneratedModelDefine": "<codeGeneratedModelDefine>\n\t{{.modelDefine}}\n\t// </codeGeneratedModelDefine>",
		"codeGeneratedModelInit":   "<codeGeneratedModelInit>\n\t\t{{.modelInit}}\n\t\t// </codeGeneratedModelInit>",
	}
	for key, item := range tags {
		reg := fmt.Sprintf(`<%s>([\s\S]+?)</%s>`, key, key)
		r, err := regexp.Compile(reg)
		if err != nil {
			return "", err
		}
		content = r.ReplaceAll(content, []byte(item))
	}
	return string(content), nil
}
