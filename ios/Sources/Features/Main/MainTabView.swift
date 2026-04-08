import SwiftUI

/// The authenticated app shell — a TabView with the three primary sections.
struct MainTabView: View {
    let authManager: AuthManager
    let apiClient: APIClient

    var body: some View {
        TabView {
            DashboardView(authManager: authManager, apiClient: apiClient)
                .tabItem {
                    Label("Dashboard", systemImage: "house")
                }

            CardsView(apiClient: apiClient)
                .tabItem {
                    Label("Cards", systemImage: "creditcard")
                }

            SettingsView(authManager: authManager)
                .tabItem {
                    Label("Settings", systemImage: "gearshape")
                }
        }
    }
}
