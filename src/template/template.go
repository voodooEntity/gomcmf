package template

import (
	"errors"
	"fmt"
	"github.com/voodooEntity/gomcmf/src/converter"
	"github.com/voodooEntity/gomcmf/src/types"
	"github.com/voodooEntity/gomcmf/src/util"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func GetNextSequence(directory string) int {
	sequences := getAllSequences(directory)
	if 0 == len(sequences) {
		return 1
	}
	highest := util.GetHighestIntValFromArray(sequences)
	return highest + 1
}

func BuildFileName(dirtyName string, sequence int, ptype string) string {
	//base64Name := base64.URLEncoding.EncodeToString([]byte(dirtyName))
	//fileName := strconv.Itoa(sequence) + "-" + base64Name + "." + ptype
	fileName := strconv.Itoa(sequence) + "." + strings.ReplaceAll(dirtyName, ".", "") + "." + ptype
	return fileName
}

func getAllSequences(directory string) []int {
	var sequences []int

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		util.Error(err.Error())
	}
	for _, file := range files {
		if file.IsDir() || !hasAllowedExt(file.Name(), GetAllowedTemplateExt()) {
			continue
		}
		fileSequence := GetSequenceFromFilename(file.Name())
		if -1 != fileSequence {
			sequences = append(sequences, fileSequence)
		}
	}
	return sequences
}

func GetNonTemplateFiles(directory string) []string {
	var files []string
	allFiles, err := ioutil.ReadDir(directory)
	if err != nil {
		util.Error(err.Error())
	}
	for _, file := range allFiles {
		if file.IsDir() || hasAllowedExt(file.Name(), GetAllowedTemplateExt()) {
			continue
		}
		files = append(files, file.Name())
	}
	return files
}

func GetAllTemplateFiles(directory string) []types.Page {
	var pageFiles []types.Page
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		util.Error(err.Error())
	}
	for _, file := range files {
		if file.IsDir() || !hasAllowedExt(file.Name(), GetAllowedTemplateExt()) {
			continue
		}
		pageFiles = append(pageFiles, GetPageByPathAndFilename(directory, file.Name()))
	}
	return pageFiles
}

func GetPageByPathAndFilename(path string, filename string) types.Page {
	_, name, ext := DecodeFileName(filename)
	urlSafeName := GetUrlSafeName(filename)
	fullPath := path + "/" + filename
	page := types.Page{
		Filename: filename,
		UrlName:  urlSafeName,
		Path:     path,
		Name:     name,
		Type:     ext,
		Sequence: GetSequenceFromFilename(filename),
		Content:  util.ReadFile(fullPath),
	}
	return page
}

func RenderPage(
	page types.Page,
	mainTemplate string,
	mainTemplateReplacements []types.Replacement,
	variables map[string]string,
	pageGroups map[string]types.Pagegroup,
	groupIdent string,
) string {
	// prestore content
	pageContent := page.Content

	// if its md we render it
	if "md" == page.Type {
		tmp := converter.Content{
			Md: pageContent,
		}
		tmp.Convert()
		// overwrite content
		pageContent = tmp.Html
	}

	pageReplacements, err := GetReplacementMarkers(pageContent)
	if nil != err {
		util.Error("Error parsing page '" + page.Name + "' for markers - error: '" + err.Error() + "'")
	}

	// replace all markers in template
	for _, pageReplacement := range pageReplacements {
		value := GetReplacementContent(pageReplacement, variables, pageGroups, "", page, groupIdent)
		pageContent = strings.ReplaceAll(pageContent, "{{"+pageReplacement.Target+"}}", value)
	}

	// now replace markers in main template
	finalPage := mainTemplate
	for _, mainTemplateReplacement := range mainTemplateReplacements {
		value := GetReplacementContent(mainTemplateReplacement, variables, pageGroups, pageContent, page, groupIdent)
		finalPage = strings.ReplaceAll(finalPage, "{{"+mainTemplateReplacement.Target+"}}", value)
	}

	return finalPage
}

func GetReplacementContent(
	replacement types.Replacement,
	variables map[string]string,
	pageGroups map[string]types.Pagegroup,
	content string,
	currPage types.Page,
	groupIdent string,
) string {
	switch replacement.Type {
	case "nav":
		// If the requested pagegroup exists
		val, ok := pageGroups[replacement.Value]
		if !ok {
			util.Error("Tryied to render non existing pagegroup '" + replacement.Value + "'")
		}
		return BuildPageGroupNav(val, replacement.Indents, currPage, groupIdent)
	case "var":
		val, ok := variables[replacement.Value]
		if !ok {
			util.Error("Tryied to render non existing variable '" + replacement.Value + "'")
		}
		return val
	case "render":
		if "content" == replacement.Value {
			return content
		}
	}
	util.Error("Unknown replacment type '" + replacement.Type + "' given")
	return ""
}

func BuildPageGroupNav(pagegroup types.Pagegroup, indents int, currPage types.Page, currIdent string) string {
	nav := ""
	if 0 < len(pagegroup.Entries) {
		spacing := ""
		if 0 < indents {
			spacing = strings.Repeat(" ", indents)
		}
		nav = nav + "<ul>"
		for _, page := range pagegroup.Entries {
			if "link" == page.Type {
				nav = nav + "\n" + spacing + "  <li><a href='" + page.Content + "' target='_blank'>" + page.Name + "</a></li>"
			} else {
				active := ""
				if currPage.Filename == page.Filename && pagegroup.Ident == currIdent {
					active = " class='active'"
				}
				nav = nav + "\n" + spacing + "  <li" + active + "><a href='" + buildInternalUrl(pagegroup.Ident, page) + "'>" + page.Name + "</a></li>"
			}
		}
		nav = nav + "\n" + spacing + "</ul>"
	}
	return nav
}

func buildInternalUrl(ident string, page types.Page) string {
	urlPath := ""
	if "/" != ident {
		urlPath = ident + "/"
	}
	fullUrl := strings.TrimPrefix(urlPath, "/") + page.UrlName + ".html"
	return fullUrl
}

func GetSequenceFromFilename(filename string) int {
	strSequence, _, _ := DecodeFileName(filename)
	seq, err := strconv.Atoi(strSequence)
	if err == nil {
		return seq
	}
	return -1
}

func GetUrlSafeName(filename string) string {
	_, name, _ := DecodeFileName(filename)
	pattern := `[^a-zA-Z0-9]+`
	regex := regexp.MustCompile(pattern)
	return regex.ReplaceAllString(name, "_")
}

func DecodeFileName(fileName string) (string, string, string) {
	fmt.Printf("filename before split %+v", fileName)
	parts := strings.Split(fileName, ".")
	fmt.Printf("filename split %+v", fileName)
	if 3 != len(parts) {
		util.Error("Invalid filename provided '" + fileName + "'")
	}
	//decodedString, err := base64.URLEncoding.DecodeString(parts[1])
	//if nil != err {
	//	util.Error("Could not decode pagename '" + parts[1] + "'")
	//}
	return parts[0], parts[1], parts[2]
}

func hasAllowedExt(filename string, allowedExts []string) bool {
	ext := filepath.Ext(filename)
	ext = strings.TrimPrefix(ext, ".")

	if util.StringInArray(allowedExts, ext) {
		return true
	}

	return false
}

func GetAllowedTemplateExt() []string {
	var allowedExts []string
	allowedExts = append(allowedExts, []string{"md", "html", "link"}...)
	return allowedExts
}

func GetReplacementMarkers(str string) ([]types.Replacement, error) {
	var replacements []types.Replacement
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		startIndex := 0
		for {
			openIndex := strings.Index(line[startIndex:], "{{")
			if openIndex == -1 {
				break
			}
			startIndex += openIndex + 2
			endIndex := strings.Index(line[startIndex:], "}}")
			if endIndex == -1 {
				return nil, errors.New("Invalid syntax: opening delimiter '{{' found without closing delimiter '}}'")
			}
			endIndex += startIndex
			content := line[startIndex:endIndex]
			replacementArray := strings.Split(content, ":")
			if len(replacementArray) < 2 {
				return nil, errors.New("Invalid syntax: replacementArray must have at least 2 entries")
			}
			indentCount := countIndent(line[:startIndex-2], startIndex-2)
			replacement := types.Replacement{
				Type:    replacementArray[0],
				Value:   replacementArray[1],
				Indents: indentCount,
				Target:  content,
			}
			if len(replacementArray) > 2 {
				replacement.Options = replacementArray[2:]
			}
			replacements = append(replacements, replacement)
			startIndex = endIndex + 2
		}
	}
	return replacements, nil
}

func countIndent(str string, startIndex int) int {
	count := 0
	for i := startIndex - 1; i >= 0; i-- {
		if str[i] == ' ' {
			count++
		} else if str[i] == '\t' {
			count += 4
		} else {
			break
		}
	}
	return count
}
