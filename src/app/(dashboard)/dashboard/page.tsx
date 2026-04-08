import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function DashboardPage() {
  return (
    <>
      <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>

      <div className="mt-6 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Active Cards</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold text-gray-900">—</p>
            <p className="mt-1 text-sm text-gray-500">Cards being tracked</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Total Credits</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold text-gray-900">—</p>
            <p className="mt-1 text-sm text-gray-500">Available credits</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Status</CardTitle>
          </CardHeader>
          <CardContent>
            <Badge variant="info">Connecting...</Badge>
            <p className="mt-2 text-sm text-gray-500">
              Waiting for API backend
            </p>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
