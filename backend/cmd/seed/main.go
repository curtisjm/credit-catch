package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"creditcatch/backend/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SeedFile struct {
	Cards []SeedCard `json:"cards"`
}

type SeedCard struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Issuer    string       `json:"issuer"`
	Network   string       `json:"network"`
	AnnualFee int          `json:"annual_fee"` // dollars in JSON
	Credits   []SeedCredit `json:"credits"`
}

type SeedCredit struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	ValueAnnual   int      `json:"value_annual"`
	Frequency     string   `json:"frequency"`
	PerPeriodMax  int      `json:"per_period_max"`
	Scope         []string `json:"scope"`
	Condition     string   `json:"condition"`
	Note          string   `json:"note"`
	AutoApply     bool     `json:"auto_apply"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	seedPath := "shared/seed/card_catalog.json"
	if len(os.Args) > 1 {
		seedPath = os.Args[1]
	}

	data, err := os.ReadFile(seedPath)
	if err != nil {
		// Try from backend/ directory
		data, err = os.ReadFile("../shared/seed/card_catalog.json")
		if err != nil {
			log.Fatalf("reading seed file: %v", err)
		}
	}

	var seed SeedFile
	if err := json.Unmarshal(data, &seed); err != nil {
		log.Fatalf("parsing seed file: %v", err)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	if err := loadSeed(ctx, pool, seed); err != nil {
		log.Fatalf("loading seed data: %v", err)
	}

	fmt.Println("Seed data loaded successfully")
}

func loadSeed(ctx context.Context, pool *pgxpool.Pool, seed SeedFile) error {
	for _, card := range seed.Cards {
		cardID, err := upsertCard(ctx, pool, card)
		if err != nil {
			return fmt.Errorf("upserting card %s: %w", card.Name, err)
		}

		for _, credit := range card.Credits {
			if err := upsertCredit(ctx, pool, cardID, credit); err != nil {
				return fmt.Errorf("upserting credit %s for %s: %w", credit.Name, card.Name, err)
			}
		}

		fmt.Printf("  ✓ %s %s (%d credits)\n", card.Issuer, card.Name, len(card.Credits))
	}

	return nil
}

func upsertCard(ctx context.Context, pool *pgxpool.Pool, card SeedCard) (string, error) {
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO card_catalog (issuer, name, network, annual_fee)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (issuer, name) DO UPDATE SET
			network = EXCLUDED.network,
			annual_fee = EXCLUDED.annual_fee,
			updated_at = now()
		RETURNING id
	`, card.Issuer, card.Name, card.Network, card.AnnualFee*100).Scan(&id)
	return id, err
}

func upsertCredit(ctx context.Context, pool *pgxpool.Pool, cardID string, credit SeedCredit) error {
	amountCents := creditAmountCents(credit)
	period := mapPeriod(credit.Frequency)
	category := strings.Join(credit.Scope, ",")
	description := credit.Condition
	if description == "" {
		description = credit.Note
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO credit_definitions (card_catalog_id, name, description, amount_cents, period, category)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING
	`, cardID, credit.Name, description, amountCents, period, category)
	return err
}

func creditAmountCents(c SeedCredit) int {
	if c.PerPeriodMax > 0 {
		return c.PerPeriodMax * 100
	}
	return c.ValueAnnual * 100
}

func mapPeriod(frequency string) string {
	switch frequency {
	case "monthly":
		return "monthly"
	case "annual":
		return "annual"
	case "semi_annual", "quarterly", "per_use":
		// Approximate as annual for now — credit period generation
		// will handle the actual cadence in Phase 2
		return "annual"
	default:
		return "annual"
	}
}
