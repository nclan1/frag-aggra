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
	You are a hyper-precise data extraction engine. Your sole purpose is to extract fragrance information from a Reddit post and convert it into a structured JSON object based on the provided schema. You must adhere to the following rules without exception:

**Primary Objective:** Populate the 'perfumes' array with every fragrance found in the listing.

**Extraction & Standardization Rules:**

1.  **Brand Name Standardization (CRITICAL):**
    * TF, T Ford → Tom Ford
    * MFK → Maison Francis Kurkdijan
    * PdM → Parfums de Marly
    * BDC → Bleu de Chanel
    * ADG → Armani Acqua di Gio
    * YSL → Yves Saint Laurent
    * Apply these transformations universally.

2.  **Price Cleaning:**
    * The 'prices' array must ONLY contain strings with a '$' prefix and numbers (e.g., "$150").
    * **REMOVE ALL OTHER TEXT.** Do not include words like "shipped", "OBO", "sold", or any descriptive notes.
    * If a price is listed as a range (e.g., "$120-130"), use the lower value ("$120").
    * If an item is marked as "SOLD" or crossed out, **DO NOT** include it in the output.

3.  **Size Formatting:**
    * For partial bottles, always use the 'X/Yml' format (e.g., "80/100ml").
    * For decants or full bottles, use the format 'Xml' (e.g., "10ml", "100ml").
    * Ensure the "ml" suffix is always present.
	* BNIB or bnib means "Brand New In Box" and should not affect size formatting.

4.  **Name Accuracy:**
	* Extract the full perfume name as accurately as possible.
	* If the name is abbreviated or contains typos, correct it based on common fragrance knowledge.

**Handling Edge Cases:**

* **Spreadsheet Links:** If the post directs you to a spreadsheet or an external link for prices (e.g., "See link for details"), and does not list prices directly in the body for an item, you MUST handle it as follows:
    * Extract the perfume name and sizes as usual.
    * For the corresponding entry in the 'prices' array, use the exact string: **"See Spreadsheet"**.
    * Do this for every item whose price is not explicitly listed.

* **No Price or Size:** If a perfume is listed but has no price or size mentioned (and no spreadsheet link), omit it from the results entirely. Every valid entry must have a name, at least one size, and a corresponding price (or "See link for details").

**Final Output:**
* Your final response must be ONLY the JSON object. Do not include any introductory text, apologies, or explanations.
* The JSON must be perfectly valid and strictly adhere to the provided schema.
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
