"use client";

import { useEffect, useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { api, ApiError } from "@/lib/api";
import type {
  CatalogCard,
  CatalogCardDetail,
  UserCard,
  Paginated,
} from "@/types/api";

function formatDollars(cents: number): string {
  return `$${(cents / 100).toFixed(0)}`;
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
      setError(err instanceof ApiError ? `Error: ${err.status}` : "Failed to load cards");
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
      const detail = await api.get<CatalogCardDetail>(`/api/v1/cards/${card.id}`);
      setAddingCard(detail);
      setNickname("");
      setStatementDay("1");
      setAddError("");
      setDialogOpen(true);
    } catch {
      setError("Failed to load card details");
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

  const handleRemove = async (userCardId: string) => {
    try {
      await api.delete(`/api/v1/me/cards/${userCardId}`);
      setUserCards((prev) => prev.filter((uc) => uc.id !== userCardId));
    } catch {
      setError("Failed to remove card");
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  return (
    <>
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">Cards</h1>
      </div>

      {error && (
        <div className="mt-4 rounded-lg bg-destructive/15 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* User's cards */}
      {userCards.length > 0 && (
        <>
          <h2 className="mt-6 text-lg font-semibold text-foreground">My Cards</h2>
          <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {userCards.map((uc) => (
              <Card key={uc.id}>
                <CardHeader className="pb-3">
                  <div className="flex items-start justify-between">
                    <div>
                      <CardTitle className="text-base">{uc.card.name}</CardTitle>
                      <p className="text-sm text-muted-foreground">
                        {capitalize(uc.card.issuer)}
                      </p>
                    </div>
                    <Badge variant="secondary">{uc.card.network}</Badge>
                  </div>
                </CardHeader>
                <CardContent>
                  {uc.nickname && (
                    <p className="text-sm text-muted-foreground">&ldquo;{uc.nickname}&rdquo;</p>
                  )}
                  <div className="mt-2 flex items-center justify-between">
                    <span className="text-sm text-muted-foreground">
                      {formatDollars(uc.card.annual_fee)}/yr fee
                    </span>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-destructive hover:text-destructive"
                      onClick={() => handleRemove(uc.id)}
                    >
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
      <h2 className="text-lg font-semibold text-foreground">
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
              className={owned ? "opacity-50" : ""}
            >
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div>
                    <CardTitle className="text-base">{card.name}</CardTitle>
                    <p className="text-sm text-muted-foreground">
                      {capitalize(card.issuer)}
                    </p>
                  </div>
                  <Badge variant="secondary">{card.network}</Badge>
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">
                    {formatDollars(card.annual_fee)}/yr
                  </span>
                  {owned ? (
                    <Badge variant="success">Added</Badge>
                  ) : (
                    <Button
                      size="sm"
                      onClick={() => handleAddClick(card)}
                    >
                      Add Card
                    </Button>
                  )}
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              Add {addingCard?.name}
            </DialogTitle>
          </DialogHeader>
          {addingCard && (
            <>
              {addingCard.credits.length > 0 && (
                <div className="mb-4">
                  <p className="text-sm font-medium text-muted-foreground mb-2">
                    Credits you&apos;ll track:
                  </p>
                  <div className="space-y-1">
                    {addingCard.credits.map((c) => (
                      <div key={c.id} className="flex justify-between text-sm">
                        <span>{c.name}</span>
                        <span className="text-primary font-medium">
                          {formatDollars(c.amount_cents)}/{c.period === "monthly" ? "mo" : "yr"}
                        </span>
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
                <Button type="submit" disabled={addLoading} className="w-full">
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
