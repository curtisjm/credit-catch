/** Matches the Go backend JSON responses exactly. */

// Card catalog (public)
export interface CatalogCard {
  id: string;
  issuer: string;
  name: string;
  network: string;
  annual_fee: number; // cents
  image_url: string;
  active: boolean;
}

export interface CreditDefinition {
  id: string;
  name: string;
  description: string;
  amount_cents: number;
  period: "monthly" | "annual" | "one_time";
  category: string;
}

export interface CatalogCardDetail extends CatalogCard {
  credits: CreditDefinition[];
}

// User cards
export interface UserCard {
  id: string;
  card: CatalogCard;
  nickname: string;
  opened_date: string | null;
  annual_fee_date: string | null;
  statement_close_day: number | null;
  active: boolean;
  created_at: string;
}

export interface AddUserCardRequest {
  card_catalog_id: string;
  nickname?: string;
  opened_date?: string;
  annual_fee_date?: string;
  statement_close_day: number;
}

// Credit periods
export interface CreditPeriod {
  id: string;
  user_card_id: string;
  credit_definition: CreditDefinition;
  period_start: string;
  period_end: string;
  used: boolean;
  used_at: string | null;
  amount_used_cents: number;
  transaction_id: string | null;
}

// Current credits (grouped by card)
export interface CurrentCardCredits {
  user_card: UserCard;
  credits: CreditPeriod[];
}

// Dashboard
export interface CardSummary {
  user_card_id: string;
  card_name: string;
  issuer: string;
  available_cents: number;
  used_cents: number;
  credits_used: number;
  credits_total: number;
}

export interface DashboardSummary {
  total_available_cents: number;
  total_used_cents: number;
  total_unused_cents: number;
  credits_used: number;
  credits_total: number;
  cards: CardSummary[];
}

export interface MonthlyDashboard {
  year: number;
  month: number;
  total_available_cents: number;
  total_used_cents: number;
  cards: CardSummary[];
}

// Paginated response wrapper
export interface Paginated<T> {
  data: T[];
  next_cursor?: string;
}
