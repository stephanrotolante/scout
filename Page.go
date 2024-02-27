package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

type Page struct {
	PageName       string
	PagePath       string
	DependentFiles []string
}

/*
Unsupported:

default require

import type
*/

var COMMENT = regexp.MustCompile(`/\*[\s\S]*?\*/|//.*`)

var DYNAMIC_IMPORT = regexp.MustCompile(`import\((?:.|\s)*?'(\..+)'`)

var DECONSTRUCTED_IMPORT = regexp.MustCompile(`import\s*{\s*(?:[A-Za-z](?:[A-Za-z0-9_])*|\s+|,)+\s*}\s*from\s*'\..+'`)

var DEFAULT_IMPORT = regexp.MustCompile(`import\s+[A-Za-z](?:[A-Za-z0-9_])*\s+from\s*'\..+'`)

var EXPORT_IMPORT = regexp.MustCompile(`export\s*{(?:[A-Za-z](?:[A-Za-z0-9_])*|,|\s)+}\s*from\s*'\..+'`)

var EXPORT_IMPORT_DEFAULT = regexp.MustCompile(`export\s+{\s*.*default.+\s*}\s*from\s*'(\..+)'`)

var IMPORT_STAR = regexp.MustCompile(`import\s+\*\s+as\s+([A-Za-z](?:[A-Za-z0-9_])*)\s+from\s+'(\..+)'`)

var DECONSTRUCTED_REQUIRED_IMPORT = regexp.MustCompile(`const\s*{\s*((?:[A-Za-z](?:[A-Za-z0-9_])*|\s|,|:)*)}\s*=\s*require\('(\..+)'\)`)

var FROM = regexp.MustCompile(`from\s*'(\..+)'`)

func (p *Page) IsFileADep(filePath string) bool {

	for _, fileDep := range p.DependentFiles {
		if strings.Contains(filePath, fileDep) {
			return true
		}
	}

	return false
}

func (p *Page) FileDepTree() {

	file, err := os.ReadFile(p.PagePath)

	if err != nil {
		log.Panic(err)
	}

	fileDirPath := GetFileDirPath(p.PagePath)

	var importedModules []Import

	fileWithoutComments := COMMENT.ReplaceAllString(string(file), "")

	// DECONSTRUCTED_IMPORT
	matches := DECONSTRUCTED_IMPORT.FindAllStringSubmatch(fileWithoutComments, -1)
	for _, match := range matches {

		stringWithoutNewLine := strings.Replace(match[0], "\n", " ", -1)

		stringWithoutDoubleSpace := strings.Replace(stringWithoutNewLine, "  ", "", -1)

		stringWithoutCommas := strings.Replace(stringWithoutDoubleSpace, ",", "", -1)

		stringSplit := strings.Fields(stringWithoutCommas)

		var startExtraction bool
		for i := 0; i < len(stringSplit); i++ {
			if stringSplit[i] == "}" {
				startExtraction = false
			}

			if startExtraction {
				importedModules = append(importedModules, Import{
					Path:   CleanFilePath(fileDirPath, strings.Replace(stringSplit[len(stringSplit)-1], "'", "", -1)),
					Module: stringSplit[i],
				})
				if stringSplit[i+1] == "as" {
					i += 2
				}
			}

			if stringSplit[i] == "{" {
				startExtraction = true
			}
		}
	}

	// DEFAULT_IMPORT
	matches = DEFAULT_IMPORT.FindAllStringSubmatch(fileWithoutComments, -1)

	for _, match := range matches {

		stringSplit := strings.Fields(match[0])

		importedModules = append(importedModules, Import{
			Path:            CleanFilePath(fileDirPath, strings.Replace(stringSplit[len(stringSplit)-1], "'", "", -1)),
			Module:          stringSplit[1],
			IsDefaultImport: true,
		})
	}

	// IMPORT_STAR
	matches = IMPORT_STAR.FindAllStringSubmatch(fileWithoutComments, -1)

	/// caputed text, import name, from
	for _, match := range matches {
		importedModules = append(importedModules, Import{
			Path:            CleanFilePath(fileDirPath, match[2]),
			Module:          match[1],
			IsDefaultImport: true,
		})
	}

	// DECONSTRUCTED_REQUIRED_IMPORT
	matches = DECONSTRUCTED_REQUIRED_IMPORT.FindAllStringSubmatch(fileWithoutComments, -1)
	for _, match := range matches {
		modulesWithoutComma := strings.ReplaceAll(match[1], ",", "")

		separatedModules := strings.Split(modulesWithoutComma, "\n")

		for _, separatedModuleImport := range separatedModules {
			if strings.Contains(separatedModuleImport, ":") {

				separatedModuleImportByColon := strings.Split(separatedModuleImport, ":")[0]

				importedModules = append(importedModules, Import{
					Path:            CleanFilePath(fileDirPath, match[2]),
					Module:          separatedModuleImportByColon,
					IsDefaultImport: false,
				})

				continue
			}
			separatedModuleImportCleaned := strings.ReplaceAll(separatedModuleImport, " ", "")
			if separatedModuleImportCleaned != "" {
				importedModules = append(importedModules, Import{
					Path:            CleanFilePath(fileDirPath, match[2]),
					Module:          separatedModuleImportCleaned,
					IsDefaultImport: false,
				})
			}

		}
	}

	//DYNAMIC_IMPORT
	matches = DYNAMIC_IMPORT.FindAllStringSubmatch(fileWithoutComments, -1)

	for _, match := range matches {
		importedModules = append(importedModules, Import{
			Path:            CleanFilePath(fileDirPath, match[1]),
			Module:          match[1],
			IsDefaultImport: true,
		})
	}

	fileModuleTracker := NewSet()

	fileTracker := NewSet()

	for {
		if len(importedModules) == 0 {

			p.DependentFiles = fileTracker.GetKeys()

			return
		}

		currentImport := importedModules[0]
		if !fileModuleTracker.Contains(fmt.Sprintf("%s:%s", currentImport.Path, currentImport.Module)) {
			fileModuleTracker.Add(fmt.Sprintf("%s:%s", currentImport.Path, currentImport.Module))
			importedModules = append(importedModules, GetDeps(&currentImport)...)

			if currentImport.FileName == "" {
				fileTracker.Add(fmt.Sprintf("%s.%s", currentImport.Path, currentImport.Extension))
			} else {
				fileTracker.Add(fmt.Sprintf("%s/%s", currentImport.Path, currentImport.FileName))
			}

		}

		importedModules = importedModules[1:]

	}

}

func GetDeps(importedModule *Import) []Import {

	fileInfo, err := os.Stat(importedModule.Path)

	var file []byte

	var correctPath string

	var importedModules []Import

	if err == nil && fileInfo.IsDir() {
		//  Folders with Index
		if file, err = os.ReadFile(path.Join(importedModule.Path, "index.js")); err == nil {
			//	fmt.Println("Found index.js")
			importedModule.SetFileName("index.js")
			correctPath = importedModule.Path
		} else if file, err = os.ReadFile(path.Join(importedModule.Path, "index.ts")); err == nil {
			// fmt.Println("Found index.ts")
			importedModule.SetFileName("index.ts")
			correctPath = importedModule.Path
		} else {
			log.Panic(fmt.Errorf("cannot find file %s", importedModule.Path))
		}

		fileWithoutComments := COMMENT.ReplaceAllString(string(file), "")

		// If modules is default import making assumption to include all files
		if !importedModule.IsDefaultImport {

			matches := EXPORT_IMPORT.FindAllStringSubmatch(fileWithoutComments, -1)

			for _, match := range matches {

				//  check if default module
				CUSTOM_DEFAULT_IMPORT_REGEX := regexp.MustCompile(fmt.Sprintf(`default\s+as\s+%s`, importedModule.Module))

				defaultMatches := CUSTOM_DEFAULT_IMPORT_REGEX.FindAllStringSubmatch(match[0], -1)

				// default module
				if len(defaultMatches) > 0 {
					fromMatches := FROM.FindAllStringSubmatch(match[0], 1)

					return []Import{
						{
							Path:            CleanFilePath(correctPath, fromMatches[0][1]),
							Module:          importedModule.Module,
							IsDefaultImport: true,
						},
					}
				}

				if strings.Contains(match[0], importedModule.Module) {

					fromMatches := FROM.FindAllStringSubmatch(match[0], 1)

					return []Import{
						{
							Path:            CleanFilePath(correctPath, fromMatches[0][1]),
							Module:          importedModule.Module,
							IsDefaultImport: false,
						},
					}
				}

			}

			return importedModules
		}

		matches := EXPORT_IMPORT_DEFAULT.FindAllStringSubmatch(fileWithoutComments, -1)

		for _, match := range matches {

			return []Import{
				{
					Path:            CleanFilePath(correctPath, match[1]),
					Module:          importedModule.Module,
					IsDefaultImport: true,
				},
			}
		}

	} else if file, err = os.ReadFile(importedModule.Path + ".js"); err == nil {
		// fmt.Println("Found JS")
		importedModule.SetExtension("js")
		correctPath = GetFileDirPath(importedModule.Path)
	} else if file, err = os.ReadFile(importedModule.Path + ".jsx"); err == nil {
		//	fmt.Println("Found JSX")
		importedModule.SetExtension("jsx")
		correctPath = GetFileDirPath(importedModule.Path)
	} else if file, err = os.ReadFile(importedModule.Path + ".ts"); err == nil {
		//	fmt.Println("Found TS")
		importedModule.SetExtension("ts")
		correctPath = GetFileDirPath(importedModule.Path)
	} else if file, err = os.ReadFile(importedModule.Path + ".tsx"); err == nil {
		// fmt.Println("Found TSX")
		importedModule.SetExtension("tsx")
		correctPath = GetFileDirPath(importedModule.Path)
	} else {
		//		log.Panic(fmt.Errorf("cannot find file %s", importedModule.Path))
		return importedModules
	}

	fileWithoutComments := COMMENT.ReplaceAllString(string(file), "")

	// DECONSTRUCTED_IMPORT
	matches := DECONSTRUCTED_IMPORT.FindAllStringSubmatch(fileWithoutComments, -1)
	for _, match := range matches {

		stringWithoutNewLine := strings.Replace(match[0], "\n", " ", -1)

		stringWithoutDoubleSpace := strings.Replace(stringWithoutNewLine, "  ", "", -1)

		stringWithoutCommas := strings.Replace(stringWithoutDoubleSpace, ",", "", -1)

		stringSplit := strings.Fields(stringWithoutCommas)

		var startExtraction bool
		for i := 0; i < len(stringSplit); i++ {
			if stringSplit[i] == "}" {
				startExtraction = false
			}

			if startExtraction {
				importedModules = append(importedModules, Import{
					Path:            CleanFilePath(correctPath, strings.Replace(stringSplit[len(stringSplit)-1], "'", "", -1)),
					Module:          stringSplit[i],
					IsDefaultImport: false,
				})
				if stringSplit[i+1] == "as" {
					i += 2
				}
			}

			if stringSplit[i] == "{" {
				startExtraction = true
			}
		}
	}

	// DEFAULT_IMPORT
	matches = DEFAULT_IMPORT.FindAllStringSubmatch(fileWithoutComments, -1)

	for _, match := range matches {

		stringSplit := strings.Fields(match[0])

		importedModules = append(importedModules, Import{
			Path:            CleanFilePath(correctPath, strings.Replace(stringSplit[len(stringSplit)-1], "'", "", -1)),
			Module:          stringSplit[1],
			IsDefaultImport: true,
		})
	}

	// IMPORT_STAR
	matches = IMPORT_STAR.FindAllStringSubmatch(fileWithoutComments, -1)

	/// caputed text, import name, from
	for _, match := range matches {
		importedModules = append(importedModules, Import{
			Path:            CleanFilePath(correctPath, match[2]),
			Module:          match[1],
			IsDefaultImport: true,
		})

	}

	// DECONSTRUCTED_REQUIRED_IMPORT
	matches = DECONSTRUCTED_REQUIRED_IMPORT.FindAllStringSubmatch(fileWithoutComments, -1)
	for _, match := range matches {
		modulesWithoutComma := strings.ReplaceAll(match[1], ",", "")

		separatedModules := strings.Split(modulesWithoutComma, "\n")

		for _, separatedModuleImport := range separatedModules {
			if strings.Contains(separatedModuleImport, ":") {

				separatedModuleImportByColon := strings.Split(separatedModuleImport, ":")[0]

				importedModules = append(importedModules, Import{
					Path:            CleanFilePath(correctPath, match[2]),
					Module:          separatedModuleImportByColon,
					IsDefaultImport: false,
				})

				continue
			}
			separatedModuleImportCleaned := strings.ReplaceAll(separatedModuleImport, " ", "")
			if separatedModuleImportCleaned != "" {
				importedModules = append(importedModules, Import{
					Path:            CleanFilePath(correctPath, match[2]),
					Module:          separatedModuleImportCleaned,
					IsDefaultImport: false,
				})
			}

		}
	}

	//DYNAMIC_IMPORT
	matches = DYNAMIC_IMPORT.FindAllStringSubmatch(fileWithoutComments, -1)

	for _, match := range matches {
		importedModules = append(importedModules, Import{
			Path:            CleanFilePath(correctPath, match[1]),
			Module:          match[1],
			IsDefaultImport: true,
		})
	}

	return importedModules
}
