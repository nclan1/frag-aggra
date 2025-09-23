package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// basically an object for each perfume, has a name, array of sizes, and array of prices
type Perfume struct {
	Name   string   `json:"name" jsonschema_description:"The standardized full brand and perfume name (e.g., 'Tom Ford Tobacco Vanille'). Apply all standardization rules."`
	Sizes  []string `json:"sizes" jsonschema_description:"An array of available sizes in ml. For partials, use 'X/Yml' format (e.g., '80/100ml')."`
	Prices []string `json:"prices" jsonschema_description:"An array of prices with '$' symbol, corresponding to each size in the sizes array."`
}

// for each listing / post, has an array of perfumes
type FragranceListing struct {
	Perfumes []Perfume `json:"perfumes" jsonschema_description:"A list of all perfumes found in the sale listing."`
}

func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

func main() {

	_ = godotenv.Load()

	systemPrompt := `
    You are a precise fragrance data extraction tool. Extract information from the provided
    Reddit perfume sale listing. Use the provided JSON schema to structure your response.
    
    Apply these critical standardization rules for brand names:
    - TF, T Ford → Tom Ford
    - MFK → Maison Francis Kurkdjian
    - PdM → Parfums de Marly
    - BDC → Bleu de Chanel
    - ADG → Armani Acqua di Gio
    - YSL → Yves Saint Laurent
    `

	redditPost := `Le Labo (Clear-Label Testers w/ Caps)

    Another 13 100ml - $230
    Another 13 50ml - $155
    Jasmin 17 100ml - $200
    Citron 28 50ml - $210
    Fleur D'Oranger 27 50ml - $130
    Le Labo Lys 41 50ml - $130

Payment Methods: Zelle & Venmo
FREE SHIPPING ON ALL BOTTLES

https://imgur.com/a/T5TfKZI`

	// define the response schema
	var FragranceListingResponseSchema = GenerateSchema[FragranceListing]()

	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)

	// create a context, cancelable, with timeout
	ctx := context.Background()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "fragrance_listing", // the name of the schema, used in the response
		Description: openai.String("Information about the perfumes extracted from a sale listing"),
		Schema:      FragranceListingResponseSchema,
		Strict:      openai.Bool(true), // if true, the model will only respond with the schema, nothing else
	}

	chat, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(redditPost),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: schemaParam,
			},
		},
		Model: openai.ChatModelGPT5Nano,
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("RAW JSON OUTPUT FROM MODEL:")
	fmt.Println(chat.Choices[0].Message.Content)

	//unmarshal the response into our struct
	var fragranceListing FragranceListing
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &fragranceListing)

	if err != nil {
		panic(err.Error())
	}

	fmt.Println("\nPARSED OUTPUT:")

	for _, perfume := range fragranceListing.Perfumes {
		fmt.Printf("Perfume: %s\n", perfume.Name)
		for i, size := range perfume.Sizes {
			price := perfume.Prices[i]
			fmt.Printf("  Size: %s - Price: %s\n", size, price)
		}
	}

}
