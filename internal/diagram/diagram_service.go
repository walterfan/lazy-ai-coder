package diagram

import (
	"fmt"
	"os"
	"strings"

	"github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/internal/models"
	"github.com/walterfan/lazy-ai-coder/internal/util"
)

// DiagramService handles diagram generation logic
type DiagramService struct{}

// NewDiagramService creates a new DiagramService
func NewDiagramService() *DiagramService {
	return &DiagramService{}
}

// DrawPlantImage generates a PlantUML diagram image
func (s *DiagramService) DrawPlantImage(imageText string, imageType string) (*models.DrawAPIResponse, error) {
	logger := log.GetLogger()
	imageScript := ""
	if imageType == "uml" {
		imageScript = util.ExtractPlantUMLScript(imageText)
	} else {
		imageScript = util.ExtractPlantMindmapScript(imageText)
	}

	if imageScript == "" {
		logger.Info("No valid PlantUML script found in the request")
		return nil, nil
	}
	plantUmlBaseUrl := os.Getenv("PLANTUML_URL")
	plantUmlPublicUrl := os.Getenv("PLANTUML_PUBLIC_URL")

	if plantUmlPublicUrl == "" {
		plantUmlPublicUrl = plantUmlBaseUrl
	}

	plantUMLClient := util.NewPlantUMLClient(plantUmlBaseUrl)
	imageUrl, _ := plantUMLClient.GeneratePngUrl(imageScript)
	imageName, _ := util.GenerateRandomString(10)
	imagePath := fmt.Sprintf("./web/images/%s.png", imageName)
	logger.Infof("Generating image script=%s, path=%s", imageScript, imagePath)
	err := plantUMLClient.GeneratePngFile(imageScript, imagePath)
	if err != nil {
		logger.Error("Failed to generate image file", "error", err)
		return nil, err
	}
	apiResponse := &models.DrawAPIResponse{
		ImageUrl:    strings.Replace(imageUrl, plantUmlBaseUrl, plantUmlPublicUrl, 1),
		ImagePath:   imagePath,
		ImageType:   imageType,
		ImageScript: imageScript,
	}
	return apiResponse, nil
}

// GenerateDiagrams generates both UML and mindmap diagrams from assistant prompt
func (s *DiagramService) GenerateDiagrams(assistantPrompt string) []models.DrawAPIResponse {
	logger := log.GetLogger()
	arrResp := []models.DrawAPIResponse{}

	if assistantPrompt == "" {
		return arrResp
	}

	// Generate UML diagram
	apiResponse, err := s.DrawPlantImage(assistantPrompt, "uml")
	if apiResponse != nil && err == nil {
		logger.Infof("Generated image: %s, URL: %s", apiResponse.ImagePath, apiResponse.ImageUrl)
		arrResp = append(arrResp, *apiResponse)
	}

	// Generate mindmap diagram
	apiResponse, err = s.DrawPlantImage(assistantPrompt, "mindmap")
	if apiResponse != nil && err == nil {
		logger.Infof("Generated image: %s, URL: %s", apiResponse.ImagePath, apiResponse.ImageUrl)
		arrResp = append(arrResp, *apiResponse)
	}

	return arrResp
}

