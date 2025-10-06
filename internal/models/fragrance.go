package models

// Perfume represents a single fragrance item for sale.
type Perfume struct {
	Name   string   `json:"name" jsonschema_description:"The standardized full brand and perfume name (e.g., 'Tom Ford Tobacco Vanille'). Apply all standardization rules."`
	Sizes  []string `json:"sizes" jsonschema_description:"An array of available sizes in ml. For partials, use 'X/Yml' format (e.g., '80/100ml')."`
	Prices []string `json:"prices" jsonschema_description:"An array of prices with '$' symbol, corresponding to each size in the sizes array."`
}

// FragranceListing represents all perfumes found in a single Reddit post.
type FragranceListing struct {
	Perfumes []Perfume `json:"perfumes" jsonschema_description:"A list of all perfumes found in the sale listing."`
}

// post raw data to pass into parser

type Post struct {
	PostID         string `json:"post_id"`
	URL            string `json:"url"`
	Title          string `json:"title"`
	Body           string `json:"body"` // The raw text to be sent to the LLM
	SellerUsername string `json:"seller_username"`
}
