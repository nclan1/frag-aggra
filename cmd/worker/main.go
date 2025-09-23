package main

import (
	"context"
	"fmt"
	"frag-aggra/internal/parser"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	p, err := parser.New()
	if err != nil {
		log.Fatalf("failed to create parser: %v", err)
	}

	redditPost := `Hi
Prices include shipping
Cashapp, Venmo, PayPal F&F
No trades
All are testers
Le Labo Santal 33 50/50ml - $130
SOLD Le Labo The Noir 29 50/50ml - $130
SOLD Le Labo Another 13 50/50ml - $130
SOLD Diptyque L'Ombre Dans L'Eau EDP 75/75ml - $120
Diptyque Eau de 34 95/100ml - $100
SOLD Diptyque Tam Dao EDP 75/75ml - $140
SOLD Diptyque Tempo EDP 75/75ml - $140`

	fmt.Println("Parsing Reddit post content...")
	listing, err := p.ParsePostContent(context.Background(), redditPost)
	if err != nil {
		log.Fatalf("failed to parse post content: %v", err)
	}
	fmt.Println("\nPARSED OUTPUT:")
	for _, perfume := range listing.Perfumes {
		fmt.Printf("Name: %s\n", perfume.Name)
		for i, size := range perfume.Sizes {
			price := perfume.Prices[i]
			fmt.Printf("  Size: %s - Price: %s\n", size, price)

		}
	}
}
