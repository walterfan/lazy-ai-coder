package util

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/walterfan/lazy-ai-coder/internal/log"
)

// PlantUMLClient defines the interface for a PlantUML generator
type PlantUMLClient interface {
	GeneratePngFile(umlText, outputPath string) error
	GeneratePngUrl(umlText string) (string, error)
}

// plantUMLClient implements the PlantUMLClient interface
type plantUMLClient struct {
	baseURL    string
	httpClient *http.Client
}

// PlantUMLOption allows setting optional parameters for PlantUMLClient
type PlantUMLOption func(*plantUMLClient)

func WithHTTPClient(client *http.Client) PlantUMLOption {
	return func(c *plantUMLClient) {
		c.httpClient = client
	}
}

// NewPlantUMLClient returns a new PlantUMLClient
func NewPlantUMLClient(baseURL string, opts ...PlantUMLOption) PlantUMLClient {
	c := &plantUMLClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
func (c *plantUMLClient) GeneratePngUrl(umlText string) (string, error) {
	encoded := encodePlantUML(umlText)

	url := fmt.Sprintf("%s/png/%s", c.baseURL, encoded)
	logger := log.GetLogger()
	logger.Infof("PlantUML image url: %s", url)
	return url, nil
}

// GeneratePngFile encodes UML text and downloads a PNG image to the given path
func (c *plantUMLClient) GeneratePngFile(umlText, outputPath string) error {

	url, err := c.GeneratePngUrl(umlText)
	if err != nil {
		return err
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// ExtractPlantUMLScript extracts the first PlantUML block from given text.
func ExtractPlantUMLScript(content string) string {
	return "@startuml\n" + ExtractPlantScript(content, "```plantuml\n@startuml", "@enduml\n```") + "\n@enduml"
}

func ExtractPlantMindmapScript(content string) string {
	return "@startmindmap\n" + ExtractPlantScript(content, "```plantuml\n@startmindmap", "@endmindmap\n```") + "\n@endmindmap"
}

func ExtractPlantScript(content, startTag, endTag string) string {
	// Look for start and end markers
	start := strings.Index(content, startTag) + len(startTag)
	if start == -1 {
		return ""
	}

	end := strings.Index(content[start:], endTag)
	if end == -1 {
		return ""
	}

	// Extract the PlantUML content including both markers
	script := content[start : start+end]

	return script
}

func encode6bit(b int) byte {
	if b < 10 {
		return byte('0' + b)
	}
	b -= 10
	if b < 26 {
		return byte('A' + b)
	}
	b -= 26
	if b < 26 {
		return byte('a' + b)
	}
	b -= 26
	if b == 0 {
		return '-'
	}
	if b == 1 {
		return '_'
	}
	return '?'
}

func append3bytes(b1, b2, b3 byte) string {
	c1 := int(b1 >> 2)
	c2 := int(((b1 & 0x3) << 4) | (b2 >> 4))
	c3 := int(((b2 & 0xF) << 2) | (b3 >> 6))
	c4 := int(b3 & 0x3F)
	return string([]byte{
		encode6bit(c1),
		encode6bit(c2),
		encode6bit(c3),
		encode6bit(c4),
	})
}

func plantumlDeflate(source string) []byte {
	var buf bytes.Buffer
	// zlib.Writer adds zlib header, we remove it manually
	zw := zlib.NewWriter(&buf)
	_, _ = zw.Write([]byte(source))
	zw.Close()

	out := buf.Bytes()
	// Strip zlib header (2 bytes) and checksum (last 4 bytes)
	return out[2 : len(out)-4]
}

func encodePlantUML(source string) string {
	compressed := plantumlDeflate(source)
	var encoded bytes.Buffer
	for i := 0; i < len(compressed); i += 3 {
		if i+2 < len(compressed) {
			encoded.WriteString(append3bytes(compressed[i], compressed[i+1], compressed[i+2]))
		} else if i+1 < len(compressed) {
			encoded.WriteString(append3bytes(compressed[i], compressed[i+1], 0))
		} else {
			encoded.WriteString(append3bytes(compressed[i], 0, 0))
		}
	}
	return encoded.String()
}
