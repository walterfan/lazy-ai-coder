package rag

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
	"github.com/unidoc/unioffice/presentation"
)

// DocumentParser handles parsing of document files
type DocumentParser struct{}

// NewDocumentParser creates a new document parser
func NewDocumentParser() *DocumentParser {
	return &DocumentParser{}
}

// ParseDocument extracts text from a document file based on its extension
func (dp *DocumentParser) ParseDocument(filePath string) (string, error) {
	info := DetectFileType(filePath)

	switch info.Ext {
	case ".pdf":
		return dp.parsePDF(filePath)
	case ".docx":
		return dp.parseDOCX(filePath)
	case ".pptx":
		return dp.parsePPTX(filePath)
	case ".md", ".rst", ".txt":
		return dp.parseText(filePath)
	default:
		return "", fmt.Errorf("unsupported document type: %s", info.Ext)
	}
}

// parsePDF extracts text from a PDF file
func (dp *DocumentParser) parsePDF(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var text strings.Builder
	totalPages := r.NumPage()

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		p := r.Page(pageNum)
		if p.V.IsNull() {
			continue
		}

		pageText, err := p.GetPlainText(nil)
		if err != nil {
			// Continue with other pages if one fails
			continue
		}

		text.WriteString(pageText)
		text.WriteString("\n\n")
	}

	content := text.String()
	if len(content) == 0 {
		return "", fmt.Errorf("no text extracted from PDF")
	}

	return content, nil
}

// parseDOCX extracts text from a DOCX file
func (dp *DocumentParser) parseDOCX(filePath string) (string, error) {
	doc, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open DOCX: %w", err)
	}
	defer doc.Close()

	docx := doc.Editable()
	content := docx.GetContent()

	if len(content) == 0 {
		return "", fmt.Errorf("no text extracted from DOCX")
	}

	return content, nil
}

// parsePPTX extracts text from a PPTX file
func (dp *DocumentParser) parsePPTX(filePath string) (string, error) {
	pres, err := presentation.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PPTX: %w", err)
	}
	defer pres.Close()

	var text strings.Builder
	slideCount := 0

	// Count slides and extract basic text
	for _, slide := range pres.Slides() {
		slideCount++
		text.WriteString(fmt.Sprintf("Slide %d:\n", slideCount))

		// Try to extract text from PlaceHolders (title, body, etc.)
		for _, ph := range slide.PlaceHolders() {
			// Try to get text content through the drawing tree
			if ph.X() != nil && ph.X().TxBody != nil {
				for _, para := range ph.X().TxBody.P {
					for _, run := range para.EG_TextRun {
						if run.R != nil && run.R.T != "" {
							text.WriteString(run.R.T)
							text.WriteString(" ")
						}
					}
					text.WriteString("\n")
				}
			}
		}

		text.WriteString("\n\n")
	}

	content := text.String()
	if len(content) == 0 {
		// Return basic info if no text extracted
		return fmt.Sprintf("PowerPoint presentation with %d slides from %s\n", slideCount, filePath), nil
	}

	return content, nil
}

// parseText reads plain text files (MD, RST, TXT)
func (dp *DocumentParser) parseText(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var text strings.Builder
	scanner := bufio.NewScanner(file)

	// Set a larger buffer for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		text.WriteString(scanner.Text())
		text.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	content := text.String()
	if len(content) == 0 {
		return "", fmt.Errorf("file is empty")
	}

	return content, nil
}

// ParseDocumentReader extracts text from an io.Reader (useful for testing)
func (dp *DocumentParser) ParseDocumentReader(r io.Reader, fileType string) (string, error) {
	// For now, only support text types from reader
	if fileType == ".md" || fileType == ".rst" || fileType == ".txt" {
		var text strings.Builder
		scanner := bufio.NewScanner(r)

		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			text.WriteString(scanner.Text())
			text.WriteString("\n")
		}

		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("error reading: %w", err)
		}

		return text.String(), nil
	}

	return "", fmt.Errorf("unsupported file type for reader: %s", fileType)
}
