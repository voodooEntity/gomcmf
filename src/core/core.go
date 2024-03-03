package core

import (
	_ "embed"
	"github.com/voodooEntity/gomcmf/src/config"
	"github.com/voodooEntity/gomcmf/src/template"
	"github.com/voodooEntity/gomcmf/src/types"
	"github.com/voodooEntity/gomcmf/src/util"
	"strconv"
	"strings"
	"time"
)

//go:embed embed/index.md
var defaultIndexFile string

//go:embed embed/404.md
var default404File string

//go:embed embed/main.html
var defaultMainFile string

//go:embed embed/config.json
var defaultConfigFile string

type Core struct {
	Command  string
	Verbose  bool
	Name     string
	Sequence int
	Type     string
	Target   string
	Input    string
	Pwd      string
}

func (self *Core) CreatePage() {
	if !util.StringInArray(template.GetAllowedTemplateExt(), self.Type) {
		util.Error("Unknown type value given '" + self.Type + "'. Allowed types are '" + strings.Join(template.GetAllowedTemplateExt(), ", ") + "'")
	}
	sequence := template.GetNextSequence(self.Pwd)
	pageName := template.BuildFileName(self.Name, sequence, self.Type)

	util.WriteFile(self.Pwd, pageName, "", false)
}

func (self *Core) BuildProject() {
	startTime := time.Now()
	pagesDirectory := config.GetValue("pagesPath")
	resourcesDirectory := config.GetValue("resourcesPath")
	variables := map[string]string{
		"base":  config.GetValue("base"),
		"title": config.GetValue("title"),
	}
	pageGroups := make(map[string]types.Pagegroup)
	outputDirectory := config.GetValue("buildPath")

	util.Print("> Building project")
	util.Print("- Current working directory: '" + self.Pwd + "'")
	util.Print("- Pages source directory: '" + pagesDirectory + "'")
	util.Print("- Output target directory: '" + outputDirectory + "'")
	util.Print("- Main template file: '" + config.GetValue("mainFile") + "'")
	util.Print("- Resources directory: '" + config.GetValue("resourcesPath") + "'")

	// read main template
	mainTemplate := util.ReadFile(self.Pwd + config.GetValue("mainFile"))
	mainTemplateReplacements, err := template.GetReplacementMarkers(mainTemplate)
	if nil != err {
		util.Error("Getting replacements for main template failed with error '" + err.Error() + "'")
	}

	// copy all files in resources recursively
	//util.CreateDirIfNotExist(self.Pwd + outputDirectory + "/" + resourcesDirectory)
	util.CopyDirectoryRecursive(self.Pwd+resourcesDirectory, self.Pwd+outputDirectory+"/"+resourcesDirectory)

	// render all pages recursive
	self.rBuildPageGroups(self.Pwd+pagesDirectory, self.Pwd+outputDirectory, "", pageGroups)

	// for each pagegroup
	for path, group := range pageGroups {
		// for each page in pagegroup
		for _, page := range group.Entries {
			// exclude link type since it doesnt need to be rendered
			if "link" != page.Type {
				pageContent := template.RenderPage(page, mainTemplate, mainTemplateReplacements, variables, pageGroups, group.Ident)
				util.CreateDirIfNotExist(self.Pwd + outputDirectory + path)
				filePath := strings.TrimPrefix(path, "/")
				if "" != filePath {
					filePath = outputDirectory + "/" + filePath
				} else {
					filePath = outputDirectory
				}
				util.WriteFile(self.Pwd+filePath, strings.TrimPrefix(page.UrlName, "/")+".html", pageContent, true)
			}
		}
	}

	// finally we render index and 404 page
	// read&render index template
	indexFile := util.ReadFile(config.GetValue("indexFile"))
	indexPage := types.Page{
		Type:     "md",
		Filename: config.GetValue("indexFile"),
		Name:     config.GetValue("title"),
		Path:     "/",
		UrlName:  "index",
		Content:  indexFile,
	}
	indexPageContent := template.RenderPage(indexPage, mainTemplate, mainTemplateReplacements, variables, pageGroups, "")
	util.WriteFile(self.Pwd+outputDirectory, indexPage.UrlName+".html", indexPageContent, true)

	// read&render 404 template
	notFoundFile := util.ReadFile(config.GetValue("404File"))
	notFoundPage := types.Page{
		Type:     "md",
		Filename: config.GetValue("404File"),
		Name:     config.GetValue("title") + " - 404",
		Path:     "/",
		UrlName:  "404",
		Content:  notFoundFile,
	}
	notFoundPageContent := template.RenderPage(notFoundPage, mainTemplate, mainTemplateReplacements, variables, pageGroups, "")
	util.WriteFile(self.Pwd+outputDirectory, notFoundPage.UrlName+".html", notFoundPageContent, true)

	elapsed := time.Since(startTime)
	util.Print("> Builded project in " + strconv.FormatInt(elapsed.Milliseconds(), 10) + " ms")
}

func (self *Core) rBuildPageGroups(pageDirectory string, outputDirectory string, currPath string, pageGroups map[string]types.Pagegroup) {
	inPath := pageDirectory + currPath
	outPath := outputDirectory + currPath
	pages := template.GetAllTemplateFiles(inPath)
	files := template.GetNonTemplateFiles(inPath)
	subDirectories := util.GetSubdirectories(inPath)

	// create all directories
	if 0 < len(subDirectories) {
		for _, subDir := range subDirectories {
			// exclude the output directory ###
			if currPath+"/"+subDir != outPath {
				//util.CreateDirIfNotExist(self.Pwd + outPath + "/" + subDir) ### disabled since it duplicates the root structure , maybe need to enable again - recheck
				self.rBuildPageGroups(pageDirectory, outputDirectory, currPath+"/"+subDir, pageGroups)
			}
		}
	}

	// copy all non-template files
	if 0 < len(files) {
		for _, file := range files {
			util.CopyFile(self.Pwd+inPath+"/"+file, self.Pwd+outPath+"/"+file)
		}
	}

	if 0 < len(pages) {
		pageGroup := types.Pagegroup{}
		for _, page := range pages {
			pageGroup.Entries = append(pageGroup.Entries, page)
		}
		if "" == currPath {
			currPath = "/"
		}
		pageGroup.Ident = currPath
		pageGroups[currPath] = pageGroup
	}
}

func (self *Core) CreateDefaultProject() {
	util.Print("> Creating default template files & directories")
	util.Print("- index.md")
	util.WriteFile(self.Pwd, "index.md", defaultIndexFile, false)
	util.Print("- 404.md")
	util.WriteFile(self.Pwd, "404.md", default404File, false)
	util.Print("- main.html")
	util.WriteFile(self.Pwd, "main.html", defaultMainFile, false)
	util.Print("- config.json")
	util.WriteFile(self.Pwd, "config.json", defaultConfigFile, false)
	util.Print("- pages/")
	util.CreateDirIfNotExist(self.Pwd + "pages")
	util.Print("- resources/")
	util.CreateDirIfNotExist(self.Pwd + "resources")
	util.Print("- output/")
	util.CreateDirIfNotExist(self.Pwd + "output")
	util.Print("")
	util.Print("> Default project files have been created")
	util.Print("For further information on usage un 'gomcmf help' or\ncheck the README on github.com/voodooEntity/gomcmf")
}
