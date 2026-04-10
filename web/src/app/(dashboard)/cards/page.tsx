"use client";

import { useEffect, useState, useCallback } from "react";
import {
  CreditCard,
  Plus,
  Trash2,
  Receipt,
  Building2,
} from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Dollars } from "@/components/dollars";
import { CardsSkeleton } from "@/components/skeletons";
import { api, ApiError } from "@/lib/api";
import { toast } from "sonner";
import type {
  CatalogCard,
  CatalogCardDetail,
  UserCard,
  Paginated,
} from "@/types/api";

const ISSUER_COLORS: Record<string, string> = {
  chase: "border-l-blue-500",
  amex: "border-l-sky-400",
  "capital one": "border-l-red-500",
  citi: "border-l-indigo-500",
  "bank of america": "border-l-red-600",
  barclays: "border-l-cyan-500",
  discover: "border-l-orange-500",
  wells_fargo: "border-l-yellow-600",
};

function issuerAccent(issuer: string): string {
  return ISSUER_COLORS[issuer.toLowerCase()] || "border-l-muted-foreground";
}

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

export default function CardsPage() {
  const [catalog, setCatalog] = useState<CatalogCard[]>([]);
  const [userCards, setUserCards] = useState<UserCard[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Add card dialog
  const [addingCard, setAddingCard] = useState<CatalogCardDetail | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [nickname, setNickname] = useState("");
  const [statementDay, setStatementDay] = useState("1");
  const [addError, setAddError] = useState("");
  const [addLoading, setAddLoading] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      const [catalogRes, userCardsRes] = await Promise.all([
        api.get<Paginated<CatalogCard>>("/api/v1/cards", {
          params: { limit: "100" },
        }),
        api.get<Paginated<UserCard>>("/api/v1/me/cards"),
      ]);
      setCatalog(catalogRes.data);
      setUserCards(userCardsRes.data);
    } catch (err) {
      setError(
        err instanceof ApiError ? `Error: ${err.status}` : "Failed to load cards",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const userCardIds = new Set(userCards.map((uc) => uc.card.id));

  const handleAddClick = async (card: CatalogCard) => {
    try {
      const detail = await api.get<CatalogCardDetail>(
        `/api/v1/cards/${card.id}`,
      );
      setAddingCard(detail);
      setNickname("");
      setStatementDay("1");
      setAddError("");
      setDialogOpen(true);
    } catch {
      toast.error("Failed to load card details");
    }
  };

  const handleAddSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!addingCard) return;
    setAddLoading(true);
    setAddError("");
    try {
      const newCard = await api.post<UserCard>("/api/v1/me/cards", {
        body: {
          card_catalog_id: addingCard.id,
          nickname: nickname || undefined,
          statement_close_day: parseInt(statementDay, 10),
        },
      });
      setUserCards((prev) => [...prev, newCard]);
      setDialogOpen(false);
      toast.success(`${addingCard.name} added`, {
        description: `Now tracking ${addingCard.credits?.length || 0} credits`,
      });
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setAddError("You already have this card");
      } else {
        setAddError("Failed to add card");
      }
    } finally {
      setAddLoading(false);
    }
  };

  const handleRemove = async (userCard: UserCard) => {
    try {
      await api.delete(`/api/v1/me/cards/${userCard.id}`);
      setUserCards((prev) => prev.filter((uc) => uc.id !== userCard.id));
      toast.success(`${userCard.card.name} removed`);
    } catch {
      toast.error("Failed to remove card");
    }
  };

  if (loading) return <CardsSkeleton />;

  const isEmpty = userCards.length === 0 && catalog.length === 0;

  return (
    <>
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">Cards</h1>
      </div>

      {error && (
        <div className="mt-4 rounded-lg bg-destructive/15 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {isEmpty ? (
        <Card className="mt-6">
          <CardContent className="flex flex-col items-center py-12 text-center">
            <CreditCard className="size-12 text-muted-foreground" />
            <h2 className="mt-4 text-lg font-semibold">No cards available</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              The card catalog is empty. Check back soon.
            </p>
          </CardContent>
        </Card>
      ) : (
        <>
          {/* User's cards */}
          {userCards.length > 0 && (
            <>
              <h2 className="mt-6 text-lg font-semibold">My Cards</h2>
              <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {userCards.map((uc) => (
                  <Card
                    key={uc.id}
                    className={`border-l-4 ${issuerAccent(uc.card.issuer)}`}
                  >
                    <CardHeader className="pb-3">
                      <div className="flex items-start justify-between">
                        <div>
                          <CardTitle className="text-base">
                            {uc.card.name}
                          </CardTitle>
                          <CardDescription className="flex items-center gap-1">
                            <Building2 className="size-3" />
                            {capitalize(uc.card.issuer)}
                          </CardDescription>
                        </div>
                        <Badge variant="secondary">{uc.card.network}</Badge>
                      </div>
                    </CardHeader>
                    <CardContent>
                      {uc.nickname && (
                        <p className="text-sm italic text-muted-foreground">
                          &ldquo;{uc.nickname}&rdquo;
                        </p>
                      )}
                      <div className="mt-2 flex items-center justify-between">
                        <span className="text-xs text-muted-foreground">
                          <Dollars
                            cents={uc.card.annual_fee}
                            showCents={false}
                            className="text-xs"
                          />
                          /yr fee
                        </span>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="gap-1 text-destructive hover:text-destructive"
                          onClick={() => handleRemove(uc)}
                        >
                          <Trash2 className="size-3.5" />
                          Remove
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
              <Separator className="my-8" />
            </>
          )}

          {/* Card catalog */}
          <h2 className="text-lg font-semibold">
            {userCards.length > 0 ? "Add More Cards" : "Browse Card Catalog"}
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Select a card to start tracking its credits
          </p>

          <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {catalog.map((card) => {
              const owned = userCardIds.has(card.id);
              return (
                <Card
                  key={card.id}
                  className={`border-l-4 ${issuerAccent(card.issuer)} ${owned ? "opacity-50" : ""}`}
                >
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                      <div>
                        <CardTitle className="text-base">{card.name}</CardTitle>
                        <CardDescription className="flex items-center gap-1">
                          <Building2 className="size-3" />
                          {capitalize(card.issuer)}
                        </CardDescription>
                      </div>
                      <Badge variant="secondary">{card.network}</Badge>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-muted-foreground">
                        <Dollars
                          cents={card.annual_fee}
                          showCents={false}
                          className="text-xs"
                        />
                        /yr
                      </span>
                      {owned ? (
                        <Badge variant="success">Added</Badge>
                      ) : (
                        <Button size="sm" onClick={() => handleAddClick(card)}>
                          <Plus className="mr-1 size-3.5" />
                          Add Card
                        </Button>
                      )}
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        </>
      )}

      {/* Add Card Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <CreditCard className="size-5 text-primary" />
              Add {addingCard?.name}
            </DialogTitle>
          </DialogHeader>
          {addingCard && (
            <>
              {addingCard.credits.length > 0 && (
                <div className="rounded-lg bg-accent p-3">
                  <p className="mb-2 flex items-center gap-1.5 text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    <Receipt className="size-3.5" />
                    Credits you&apos;ll track
                  </p>
                  <div className="space-y-1.5">
                    {addingCard.credits.map((c) => (
                      <div
                        key={c.id}
                        className="flex justify-between text-sm"
                      >
                        <span>{c.name}</span>
                        <Dollars
                          cents={c.amount_cents}
                          className="text-sm font-medium text-primary"
                        />
                      </div>
                    ))}
                  </div>
                </div>
              )}
              <form onSubmit={handleAddSubmit} className="space-y-4">
                {addError && (
                  <div className="rounded-lg bg-destructive/15 px-4 py-3 text-sm text-destructive">
                    {addError}
                  </div>
                )}
                <div className="space-y-2">
                  <Label htmlFor="nickname">Nickname (optional)</Label>
                  <Input
                    id="nickname"
                    placeholder="e.g. Personal, Business"
                    value={nickname}
                    onChange={(e) => setNickname(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="statement-day">Statement close day</Label>
                  <Input
                    id="statement-day"
                    type="number"
                    min={1}
                    max={31}
                    required
                    value={statementDay}
                    onChange={(e) => setStatementDay(e.target.value)}
                  />
                  <p className="text-xs text-muted-foreground">
                    Day of the month your statement closes (1-31)
                  </p>
                </div>
                <Button
                  type="submit"
                  disabled={addLoading}
                  className="w-full"
                >
                  {addLoading ? "Adding..." : "Add Card"}
                </Button>
              </form>
            </>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
}
