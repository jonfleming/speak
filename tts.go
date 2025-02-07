package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const (
	apiBaseURL = "https://api.sws.speechify.com"
	voiceID    = "bwyneth"
	filePath   = "./audio.mp3"
	maxLen     = 2500
)

var apiKey string

type AudioRequest struct {
	Input       string `json:"input"`
	VoiceID     string `json:"voice_id"`
	AudioFormat string `json:"audio_format"`
}

type AudioResponse struct {
	AudioData string `json:"audio_data"`
}

func getAudio(text string) ([]byte, error) {
	requestBody, err := json.Marshal(AudioRequest{
		Input:       fmt.Sprintf("<speak>%s</speak>", text),
		VoiceID:     voiceID,
		AudioFormat: "mp3",
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/audio/speech", apiBaseURL), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%d %s\n%s", resp.StatusCode, resp.Status, string(body))
	}

	var audioResponse AudioResponse
	if err := json.NewDecoder(resp.Body).Decode(&audioResponse); err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(audioResponse.AudioData)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file", err)
	}

	apiKey = os.Getenv("API_KEY")
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run tts.go <filename>")
		os.Exit(1)
	}

	filename := os.Args[1]
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	text := string(fileContent)
	if len(text) == 0 {
		fmt.Println("No text provided in file.")
		os.Exit(1)
	}

	if len(text) > maxLen {
		text = text[:maxLen]
	}

	fmt.Println("Text: ", text)
	audioData, err := getAudio(text)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(filePath, audioData, 0644); err != nil {
		fmt.Println("Error writing file:", err)
		os.Exit(1)
	}

	fmt.Println("Audio file saved as", filePath)
}
