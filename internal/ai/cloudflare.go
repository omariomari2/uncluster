package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type CloudflareConfig struct {
	AccountID string
	APIToken  string
	Model     string
	Enabled   bool
}

type CloudflareClient struct {
	config     CloudflareConfig
	httpClient *http.Client
}

func NewCloudflareClient(config CloudflareConfig) *CloudflareClient {
	if config.Model == "" {
		config.Model = "@cf/meta/llama-3-8b-instruct"
	}

	return &CloudflareClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type AIAnalysisRequest struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AIAnalysisResponse struct {
	Result struct {
		Response string `json:"response"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"errors"`
}

type ComponentAnalysisResult struct {
	ShouldBeComponent bool     `json:"shouldBeComponent"`
	Reason            string   `json:"reason"`
	ComponentName     string   `json:"componentName"`
	Props             []string `json:"props"`
	Pattern           string   `json:"pattern"`
	Confidence        string   `json:"confidence"` // "high", "medium", "low"
}

func (c *CloudflareClient) AnalyzeHTMLForComponents(htmlContent string, elementInfo string) (*ComponentAnalysisResult, error) {
	if !c.config.Enabled {
		return nil, fmt.Errorf("Cloudflare AI is not enabled")
	}

	if c.config.AccountID == "" || c.config.APIToken == "" {
		return nil, fmt.Errorf("Cloudflare credentials not configured")
	}

	prompt := c.buildComponentAnalysisPrompt(htmlContent, elementInfo)

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/%s",
		c.config.AccountID, c.config.Model)

	requestBody := AIAnalysisRequest{
		Messages: []Message{
			{
				Role: "system",
				Content: `You are an expert React component architect. Your job is to analyze HTML elements and determine if they should become React components.

Rules:
1. NOT every div should be a component - only meaningful, reusable patterns
2. Components should have semantic meaning (cards, buttons, forms, navigation items, etc.)
3. Generic wrapper divs, layout containers, or single-use elements should NOT be components
4. Look for patterns that appear multiple times OR have significant structure
5. Consider if the element would benefit from props and reusability

Respond with a JSON object containing:
- shouldBeComponent: boolean
- reason: string explaining why or why not
- componentName: string (if shouldBeComponent is true, suggest a PascalCase name)
- props: array of suggested prop names (if applicable)
- pattern: string describing the pattern (e.g., "card", "button", "form-field")
- confidence: "high", "medium", or "low"`,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIToken)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("🤖 Sending HTML element to Cloudflare AI for analysis...")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ Cloudflare AI returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("Cloudflare AI API returned status %d", resp.StatusCode)
	}

	var aiResponse AIAnalysisResponse
	if err := json.Unmarshal(body, &aiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !aiResponse.Success {
		if len(aiResponse.Errors) > 0 {
			return nil, fmt.Errorf("Cloudflare AI error: %s", aiResponse.Errors[0].Message)
		}
		return nil, fmt.Errorf("Cloudflare AI request failed")
	}

	result, err := c.parseAIResponse(aiResponse.Result.Response)
	if err != nil {
		log.Printf("⚠️ Failed to parse AI response as JSON, using fallback: %v", err)
		result = &ComponentAnalysisResult{
			ShouldBeComponent: strings.Contains(strings.ToLower(aiResponse.Result.Response), "shouldbecomponent: true") ||
				strings.Contains(strings.ToLower(aiResponse.Result.Response), "\"shouldbecomponent\": true"),
			Reason:     aiResponse.Result.Response,
			Confidence: "medium",
		}
	}

	log.Printf("✅ AI analysis complete: shouldBeComponent=%v, confidence=%s", result.ShouldBeComponent, result.Confidence)

	return result, nil
}

func (c *CloudflareClient) buildComponentAnalysisPrompt(htmlContent string, elementInfo string) string {
	maxHTMLLength := 2000
	if len(htmlContent) > maxHTMLLength {
		htmlContent = htmlContent[:maxHTMLLength] + "... [truncated]"
	}

	return fmt.Sprintf(`Analyze this HTML element and determine if it should become a React component:

Element Information:
%s

HTML Content:
%s

Provide your analysis as a JSON object with the fields: shouldBeComponent, reason, componentName, props, pattern, and confidence.`, elementInfo, htmlContent)
}

func (c *CloudflareClient) parseAIResponse(responseText string) (*ComponentAnalysisResult, error) {
	responseText = strings.TrimSpace(responseText)

	if strings.HasPrefix(responseText, "```json") {
		responseText = strings.TrimPrefix(responseText, "```json")
		responseText = strings.TrimSuffix(responseText, "```")
	} else if strings.HasPrefix(responseText, "```") {
		responseText = strings.TrimPrefix(responseText, "```")
		responseText = strings.TrimSuffix(responseText, "```")
	}

	startIdx := strings.Index(responseText, "{")
	endIdx := strings.LastIndex(responseText, "}")

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return nil, fmt.Errorf("no JSON object found in response")
	}

	jsonStr := responseText[startIdx : endIdx+1]

	var result ComponentAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &result, nil
}

func (c *CloudflareClient) IsEnabled() bool {
	return c.config.Enabled && c.config.AccountID != "" && c.config.APIToken != ""
}
