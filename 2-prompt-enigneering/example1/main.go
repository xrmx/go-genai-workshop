package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/predictionguard/go-client"
)

// Define the API details to access the LLM.
var host = "https://api.predictionguard.com"
var apiKey = os.Getenv("PGKEY")

// qAPromptTemplate is a template for a question and answer prompt.
func qAPromptTemplate(context, question string) string {
	return fmt.Sprintf(`Read the context below and answer the question. If the question cannot be answered based on the context alone or the context does not explicitly say the answer to the question, respond "Sorry I had trouble answering this question, based on the information I found."

Context: "%s"

Question: "%s"
`, context, question)
}

func run(query string) error {

	logger := func(ctx context.Context, msg string, v ...any) {
		s := fmt.Sprintf("msg: %s", msg)
		for i := 0; i < len(v); i = i + 2 {
			s = s + fmt.Sprintf(", %s: %v", v[i], v[i+1])
		}
		log.Println(s)
	}

	cln := client.New(logger, host, apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Read in the context from a file.
	promptContext, err := os.ReadFile("context1.txt")
	if err != nil {
		log.Fatal(err)
	}

	input := []client.ChatMessage{
		{
			Role: client.Roles.User,
			Content: qAPromptTemplate(
				string(promptContext),
				query,
			),
		},
	}

	ch := make(chan client.ChatSSE, 1000)

	err = cln.ChatSSE(ctx, client.Models.Hermes2ProMistral7B, input, 1000, 0.1, ch)
	if err != nil {
		return fmt.Errorf("ERROR: %w", err)
	}

	for resp := range ch {
		for _, choice := range resp.Choices {
			fmt.Print(choice.Delta.Content)
		}
	}

	return nil
}

func main() {
	if err := run("When did we add an additional endpoint to the API?"); err != nil {
		log.Fatalln(err)
	}
}
