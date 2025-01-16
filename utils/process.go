package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	padding  = 2
	maxWidth = 80
)

type SpringProject struct {
	GroupID      string
	ArtifactID   string
	ProjectName  string
	Dependencies []string
}

func CreateProject(project *SpringProject, progress *progress.Model) tea.Cmd {
	return func() tea.Msg {
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
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error in response code: %s\n", resp.Status)
			return err
		}

		/**buf := new(bytes.Buffer)
			_, err = io.Copy(buf, resp.Body)
			if err != nil {
				fmt.Printf("Error reading the downloaded file: %v\n", err)
				return
			}

			// Unzip the file
			err = unzip(buf.Bytes(), project.ProjectName)
			if err != nil {
				fmt.Printf("Error unziping the file: %v\n", err)
				return
			}

			fmt.Printf("Project '%s' succesfully created in folder: '%s'\n", project.ProjectName, project.ProjectName)
		  **/
		progress.IncrPercent(1)

		return "updating"
	}
}

func unzip(data []byte, dest string) error {
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
