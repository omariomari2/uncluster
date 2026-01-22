package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WorkerAIConfig defines connection settings for a Workers AI-backed endpoint.
type WorkerAIConfig struct {
	URL     string
	Token   string
	Model   string
	Enabled bool
}

// WorkerAIClient uses a Cloudflare Worker to run component analysis.
type WorkerAIClient struct {
	config     WorkerAIConfig
	httpClient *http.Client
}

func NewWorkerAIClient(config WorkerAIConfig) *WorkerAIClient {
	if config.Model == "" {
		config.Model = "@cf/meta/llama-3-8b-instruct"
	}

	return &WorkerAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type workerAIRequest struct {
	HTML        string `json:"html"`
	ElementInfo string `json:"elementInfo"`
	Model       string `json:"model,omitempty"`
}

type workerAIResponse struct {
	Success bool                   `json:"success"`
	Result  *ComponentAnalysisResult `json:"result,omitempty"`
	Raw     string                 `json:"raw,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

func (c *WorkerAIClient) AnalyzeHTMLForComponents(htmlContent string, elementInfo string) (*ComponentAnalysisResult, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("Workers AI is not enabled")
	}

	payload := workerAIRequest{
		HTML:        htmlContent,
		ElementInfo: elementInfo,
		Model:       c.config.Model,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(c.config.Token) != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}

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
		return nil, fmt.Errorf("Workers AI returned status %d", resp.StatusCode)
	}

	var aiResponse workerAIResponse
	if err := json.Unmarshal(body, &aiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !aiResponse.Success {
		if aiResponse.Error != "" {
			return nil, fmt.Errorf("Workers AI error: %s", aiResponse.Error)
		}
		return nil, fmt.Errorf("Workers AI request failed")
	}

	if aiResponse.Result != nil {
		return aiResponse.Result, nil
	}

	if aiResponse.Raw != "" {
		parsed, err := parseComponentAnalysis(aiResponse.Raw)
		if err != nil {
			return nil, err
		}
		return parsed, nil
	}

	return nil, fmt.Errorf("Workers AI response missing result")
}

func (c *WorkerAIClient) IsEnabled() bool {
	return c.config.Enabled && strings.TrimSpace(c.config.URL) != ""
}

func parseComponentAnalysis(responseText string) (*ComponentAnalysisResult, error) {
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
