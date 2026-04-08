import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function SettingsPage() {
  return (
    <>
      <h1 className="text-2xl font-bold text-gray-900">Settings</h1>

      <div className="mt-6 space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Profile</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-gray-500">
              Profile settings will be available once connected to the API.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Notifications</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-gray-500">
              Notification preferences will be available once connected to the
              API.
            </p>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
