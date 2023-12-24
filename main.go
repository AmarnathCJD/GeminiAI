package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

var (
	apiKey = os.Getenv("GEMINI_API_KEY") //  GET FROM GOOGLE MAKERSUIT
)

func sendPromt(prompt string) {
	req, _ := http.NewRequest("POST", "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key="+apiKey, promptToPayload(prompt))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	var GeminiResponse GeminiResponse
	if err = json.NewDecoder(res.Body).Decode(&GeminiResponse); err != nil {
		panic(err)
	}

	print("Gemini: ")

	for _, candidate := range GeminiResponse.Candidates {
		for _, part := range candidate.Content.Parts {
			print(part.Text)
		}
	}
	print("\n")
}

func prepareNewlinelessBase64ofImageFile(path string) string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	b, _ := io.ReadAll(file)
	baseencoded := base64.StdEncoding.EncodeToString(b)
	return strings.ReplaceAll(baseencoded, "\n", "")
}

func sendImagePrompt(prompt string, file string) {
	req, _ := http.NewRequest("POST", "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro-vision:generateContent?key="+apiKey, strings.NewReader(`{
		"contents":[
		  {
			"parts":[
			  {"text": "`+prompt+`"},
			  {
				"inline_data": {
				  "mime_type":"image/jpeg",
				  "data": "`+prepareNewlinelessBase64ofImageFile(file)+`"
				}
			  }
			]
		  }
		]
	  }
	`))

	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	var GeminiResponse GeminiResponse
	if err = json.NewDecoder(res.Body).Decode(&GeminiResponse); err != nil {
		panic(err)
	}

	print("Gemini: ")

	for _, candidate := range GeminiResponse.Candidates {
		for _, part := range candidate.Content.Parts {
			print(part.Text)
		}
	}
	print("\n")
}

type GemPayload struct {
	Contents string
}

func promptToPayload(p string) io.Reader {
	var data = strings.NewReader(`{
		"contents": [{
		  "parts":[{
			"text": "` + p + `"}]}]}`)
	return data
}

func main() {
	choice := 0
	fmt.Println("Welcome to Gemini CLI")
	fmt.Println("1. Text Prompt")
	fmt.Print("2. Image Prompt\n\nEnter your choice: ")
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		fmt.Println()
		for {
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("You: ")
			prompt, _ := reader.ReadString('\n')
			prompt = strings.TrimSpace(prompt)

			sendPromt(prompt)
		}
	case 2:
		fmt.Print("\nEnter the path to the image file: ")
		var path string
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			path = scanner.Text()
		}

		fmt.Println("Resizing image to 512x512...")
		resizeImageFileTo512x512(path)

		fmt.Print("Enter the text prompt: ")
		var prompt string
		if scanner.Scan() {
			prompt = scanner.Text()
		}

		sendImagePrompt(prompt, "test_resized.jpg")
	}
}

func resizeImageFileTo512x512(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	dim_x, dim_y := getImageDimensions(path)

	m := resize.Resize(512, uint(aspectWidth(dim_x, dim_y)), img, resize.Lanczos3)

	out, err := os.Create("test_resized.jpg")
	if err != nil {
		return err
	}
	defer out.Close()
	err = jpeg.Encode(out, m, nil)
	if err != nil {
		return err
	}

	return nil
}

func aspectWidth(height int, width int) int {
	return (height / width) * 512
}

func getImageDimensions(path string) (int, int) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		panic(err)
	}

	return image.Height, image.Width
}
