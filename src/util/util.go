package util

import (
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

var loggerOut = log.New(os.Stdout, "", 0)

func Explode(delimiter, text string) []string {
	if len(delimiter) <= len(text) {
		return strings.Split(text, delimiter)
	}
	return []string{text}
}

func WriteFile(path string, file string, content string, overwrite bool) {
    fullPath := filepath.Join(path, file)
    if !overwrite {
        if _, err := os.Stat(fullPath); err == nil {
            Error("File '" + file + "' at path '" + path + "' already exists. Exiting")
        }
    }
    err := os.WriteFile(fullPath, []byte(content), 0644)
    if nil != err {
        Error(err.Error())
    }
}

func Print(text string) {
	loggerOut.Println(text)
}

func Error(text string) {
	loggerOut.Println(text)
	os.Exit(0)
}

func GetHighestIntValFromArray(input []int) int {
	if 0 == len(input) {
		return -1
	}

	highest := input[0]
	for _, value := range input {
		if value > highest {
			highest = value
		}
	}
	return highest
}

func StringInArray(haystack []string, needle string) bool {
	if 0 == len(haystack) {
		return false
	}
	for _, val := range haystack {
		if needle == val {
			return true
		}
	}
	return false
}

func GetSubdirectories(directory string) []string {
    var directories []string
    allFiles, err := os.ReadDir(directory)
    if err != nil {
        Error(err.Error())
    }
    for _, file := range allFiles {
        if !file.IsDir() {
            continue
        }
        directories = append(directories, file.Name())
    }
    return directories
}

func CreateDirIfNotExist(dir string) {
	fmt.Printf("Directory to create %+v \n", dir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			Error("Could not create directory '" + dir + "' with error '" + err.Error() + "'")
		}
	}
}

func CopyDirectoryRecursive(sourceDirectory string, targetDirectory string) error {
	// Create target directory
	CreateDirIfNotExist(targetDirectory)

	// Read source directory
	entries, err := os.ReadDir(sourceDirectory)
	if err != nil {
		return err
	}

	// Copy all files and directories in the source directory
	for _, entry := range entries {
		srcPath := filepath.Join(sourceDirectory, entry.Name())
		dstPath := filepath.Join(targetDirectory, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectories
			err = CopyDirectoryRecursive(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Copy files
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyFile(file string, target string) error {
	srcFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

func ReadFile(filepath string) string {
	data, err := os.ReadFile(filepath)
	if nil != err {
		Error("Could not read file" + err.Error())
	}
	return string(data)
}

func StringToBOSS(input string) string {
	result := ""
	for _, char := range input {
		result += fmt.Sprintf("%03d", int(char))
	}
	return result
}

func BOSSToString(input string) string {
	result := ""
	if 0 != len(input)%3 {
		Error("BOS String has invalid length '" + input + "'")
	}
	for i := 0; i < len(input); i += 3 {
		ordBlock := input[i : i+3]
		ordNr, err := strconv.Atoi(ordBlock)
		if nil != err {
			Error("Given ord block cant be converted to int '" + ordBlock + "'")
		}

		result = result + string(ordNr)
	}
	return result
}
