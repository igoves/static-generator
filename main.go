package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
)

type Page struct {
	Slug        string
	Title       string
	Description string
}

func loadCSV(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to load CSV from URL %s: %v", url, err)
	}
	defer resp.Body.Close()

	csvReader := csv.NewReader(resp.Body)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %v", err)
	}

	return records, nil
}


func loadTemplate(templatePath string) (*template.Template, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template from file %s: %v", templatePath, err)
	}
	return tmpl, nil
}


func createDistDirectory(distDir string) error {
	err := os.MkdirAll(distDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", distDir, err)
	}
	return nil
}


func generateHTML(page Page, tmpl *template.Template, distDir string, wg *sync.WaitGroup) {
	defer wg.Done()

	filePath := filepath.Join(distDir, page.Slug+".html")

	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Failed to create file %s: %v", filePath, err)
		return
	}
	defer file.Close()

	err = tmpl.Execute(file, page)
	if err != nil {
		log.Printf("Failed to write to file %s: %v", filePath, err)
		return
	}

	log.Printf("File %s created successfully", filePath)
}

func main() {
	var url, templatePath, distDir string
	var wg sync.WaitGroup

	var rootCmd = &cobra.Command{
		Use:   "static-generator",
		Short: "Generator of static pages from CSV",
		Run: func(cmd *cobra.Command, args []string) {
			records, err := loadCSV(url)
			if err != nil {
				log.Fatalf("Error loading CSV: %v", err)
			}

			var pages []Page
			for _, record := range records[1:] {
				page := Page{
					Slug:        record[0],
					Title:       record[1],
					Description: record[2],
				}
				pages = append(pages, page)
			}

			err = createDistDirectory(distDir)
			if err != nil {
				log.Fatalf("Error when creating directory: %v", err)
			}

			tmpl, err := loadTemplate(templatePath)
			if err != nil {
				log.Fatalf("Error loading template: %v", err)
			}

			for _, page := range pages {
				wg.Add(1)
				go generateHTML(page, tmpl, distDir, &wg)
			}

			wg.Wait()

			log.Println("All pages have been created successfully.")
		},
	}

	rootCmd.Flags().StringVarP(&url, "url", "u", "", "CSV download URL (required)")
	rootCmd.Flags().StringVarP(&templatePath, "template", "t", "template.html", "Path to template HTML")
	rootCmd.Flags().StringVarP(&distDir, "dist", "d", "dist", "Directory for storing HTML files")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error when executing the command: %v", err)
	}
}



//
