package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
    "os/exec"
    "runtime"

	"github.com/joho/godotenv"
)

const (
	apiBaseURL = "https://api.sws.speechify.com"
	voiceID    = "bwyneth"
	outputFile   = "./audio.mp3"
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
    execPath, err := os.Executable()
    if err != nil {
        fmt.Println("Error getting executable path:", err)
        os.Exit(1)
    }

    execDir := filepath.Dir(execPath)
    envPath := filepath.Join(execDir, ".env")

    // Try loading .env from the executable's directory
    err = godotenv.Load(envPath)
    if err != nil {
        // Fall back to loading .env from the current working directory
        err = godotenv.Load()
        if err != nil {
            fmt.Println("Error loading .env file", err)
            os.Exit(1)
        }
    }

	apiKey = os.Getenv("API_KEY")

    var text string

    // Check if input is being piped via stdin
    fileInfo, err := os.Stdin.Stat()
    if err != nil {
        fmt.Println("Error checking stdin:", err)
        os.Exit(1)
    }

    if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
        // Read from stdin
        input, err := ioutil.ReadAll(os.Stdin)
        if err != nil {
            fmt.Println("Error reading from stdin:", err)
            os.Exit(1)
        }
        text = string(input)
    } else {
        // Read from file if no stdin input
        if len(os.Args) < 2 {
            fmt.Println("Usage: speak <filename> or pipe text via stdin")
            os.Exit(1)
        }

        filename := os.Args[1]
        fileContent, err := ioutil.ReadFile(filename)
        if err != nil {
            fmt.Println("Error reading file:", err)
            os.Exit(1)
        }
        text = string(fileContent)
    }

	if len(text) == 0 {
		fmt.Println("No text provided in file.")
		os.Exit(1)
	}

	if len(text) > maxLen {
		text = text[:maxLen]
	}

	fmt.Println("Playing: ", text)
	audioData, err := getAudio(text)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

   // Write audio to a temporary file for playback
   tmpFile, err := ioutil.TempFile("", "speak-*.mp3")
   if err != nil {
       fmt.Println("Error creating temp file:", err)
       os.Exit(1)
   }
   defer os.Remove(tmpFile.Name())

   if _, err := tmpFile.Write(audioData); err != nil {
       fmt.Println("Error writing to temp file:", err)
       os.Exit(1)
   }
   if err := tmpFile.Close(); err != nil {
       fmt.Println("Error closing temp file:", err)
       os.Exit(1)
   }

   // Determine playback command based on OS
   var cmd *exec.Cmd
   switch runtime.GOOS {
   case "darwin":
       cmd = exec.Command("afplay", tmpFile.Name())
   case "linux":
       cmd = exec.Command("play", tmpFile.Name())
   case "windows":
       cmd = exec.Command("cmdmp3", tmpFile.Name())
   default:
       fmt.Println("Unsupported OS for audio playback:", runtime.GOOS)
       os.Exit(1)
   }
//    cmd.Stdout = os.Stdout
//    cmd.Stderr = os.Stderr
   cmd.Stdout = ioutil.Discard
   cmd.Stderr = ioutil.Discard

   if err := cmd.Run(); err != nil {
       fmt.Println("Error playing audio:", err)
       os.Exit(1)
   }
}
