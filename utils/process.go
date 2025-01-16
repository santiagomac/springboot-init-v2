package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type SpringProject struct {
	GroupID      string
	ArtifactID   string
	ProjectName  string
	Dependencies []string
}

func CreateProject(project *SpringProject) (data []byte, dest string, err error) {
	// URL BUILDING TO CONSUME Spring Initializr
	url := fmt.Sprintf("https://start.spring.io/starter.zip?type=gradle-project&language=java&bootVersion=3.3.0&groupId=%s&artifactId=%s&name=%s&dependencies=%s",
		project.GroupID,
		project.ArtifactID,
		project.ProjectName,
		"web,jpa",
	)

	// DOWNLOAD THE PROJECT
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading the project: %v\n", err)
		return nil, "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error in response code: %s\n", resp.Status)
		return nil, "", err
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		fmt.Printf("Error reading the downloaded file: %v\n", err)
		return nil, "", err
	}

	// Unzip the file
	/*err = unzip(buf.Bytes(), project.ProjectName)
	if err != nil {
		fmt.Printf("Error unziping the file: %v\n", err)
		return nil, "", err
	}*/

	fmt.Printf("Project '%s' succesfully created in folder: '%s'\n", project.ProjectName, project.ProjectName)
	return buf.Bytes(), project.ProjectName, nil
}

func Unzip(data []byte, dest string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(dest, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
			continue
		}

		os.MkdirAll(filepath.Dir(path), os.ModePerm)

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		defer outFile.Close()
		rc, err := file.Open()
		if err != nil {
			return err
		}

		defer rc.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}

	return nil
}
