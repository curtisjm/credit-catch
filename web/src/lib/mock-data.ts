// Mock data for development — matches API spec schemas

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

export interface CreditPeriodMock {
  id: string;
  card_name: string;
  credit_name: string;
  amount_cents: number;
  period_start: string;
  period_end: string;
  used: boolean;
  status: "used" | "expiring" | "expired" | "upcoming";
}

export const mockDashboard: DashboardSummary = {
  total_available_cents: 32500,
  total_used_cents: 18000,
  total_unused_cents: 14500,
  credits_used: 7,
  credits_total: 12,
  cards: [
    {
      user_card_id: "c1a1",
      card_name: "Sapphire Reserve",
      issuer: "Chase",
      available_cents: 15000,
      used_cents: 9000,
      credits_used: 3,
      credits_total: 5,
    },
    {
      user_card_id: "c2b2",
      card_name: "Gold Card",
      issuer: "Amex",
      available_cents: 12000,
      used_cents: 6000,
      credits_used: 3,
      credits_total: 4,
    },
    {
      user_card_id: "c3c3",
      card_name: "Venture X",
      issuer: "Capital One",
      available_cents: 5500,
      used_cents: 3000,
      credits_used: 1,
      credits_total: 3,
    },
  ],
};

export const mockCredits: CreditPeriodMock[] = [
  {
    id: "cr1",
    card_name: "Sapphire Reserve",
    credit_name: "Travel Credit",
    amount_cents: 5000,
    period_start: "2026-04-01",
    period_end: "2026-04-30",
    used: true,
    status: "used",
  },
  {
    id: "cr2",
    card_name: "Gold Card",
    credit_name: "Dining Credit",
    amount_cents: 1000,
    period_start: "2026-04-01",
    period_end: "2026-04-30",
    used: false,
    status: "expiring",
  },
  {
    id: "cr3",
    card_name: "Gold Card",
    credit_name: "Uber Credit",
    amount_cents: 1000,
    period_start: "2026-04-01",
    period_end: "2026-04-30",
    used: true,
    status: "used",
  },
  {
    id: "cr4",
    card_name: "Sapphire Reserve",
    credit_name: "DoorDash Credit",
    amount_cents: 500,
    period_start: "2026-03-01",
    period_end: "2026-03-31",
    used: false,
    status: "expired",
  },
  {
    id: "cr5",
    card_name: "Venture X",
    credit_name: "Travel Credit",
    amount_cents: 5500,
    period_start: "2026-05-01",
    period_end: "2026-05-31",
    used: false,
    status: "upcoming",
  },
];
