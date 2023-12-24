package main

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content struct {
		Parts []Part `json:"parts"`
	}
}

type Part struct {
	Text string `json:"text"`
}
