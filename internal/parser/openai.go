package parser

import (
	"context"
	"encoding/json"
	"errors"
	"frag-aggra/internal/models"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

type Parser struct {
	client       *openai.Client
	systemPrompt string
}

func New() (*Parser, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable not set")
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))

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

	return &Parser{
		client:       &client,
		systemPrompt: systemPrompt,
	}, nil

}

func (p *Parser) ParsePostContent(ctx context.Context, postContent string) (*models.FragranceListing, error) {

	var FragranceListingSchema = generateSchema[models.FragranceListing]()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "fragrance_listing",
		Description: openai.String("Information about the perfumes extracted from a sale listing"),
		Schema:      FragranceListingSchema,
		Strict:      openai.Bool(true),
	}

	resp, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(p.systemPrompt),
			openai.UserMessage(postContent),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: schemaParam,
			},
		},
		Model: openai.ChatModelGPT4o2024_08_06,
	})

	if err != nil {
		return nil, err
	}

	var listing models.FragranceListing
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &listing)
	if err != nil {
		return nil, err
	}

	return &listing, nil
}

func generateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}
