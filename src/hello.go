package main

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

func main() {

	_ = godotenv.Load()

	productInfo := `Le Labo (Clear-Label Testers w/ Caps)

    Another 13 100ml - $230
    Another 13 50ml - $155
    Jasmin 17 100ml - $200
    Citron 28 50ml - $210
    Fleur D'Oranger 27 50ml - $130
    Le Labo Lys 41 50ml - $130

Payment Methods: Zelle & Venmo
FREE SHIPPING ON ALL BOTTLES

https://imgur.com/a/T5TfKZI`
	println(productInfo)
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Say this is a test"),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		panic(err.Error())
	}
	println(chatCompletion.Choices[0].Message.Content)
}
