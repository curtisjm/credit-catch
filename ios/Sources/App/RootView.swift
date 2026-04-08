import SwiftUI

/// Root view that switches between auth and main flows based on authentication state.
struct RootView: View {
    let authManager: AuthManager
    let apiClient: APIClient

    var body: some View {
        Group {
            if authManager.isLoading {
                ProgressView("Loading...")
            } else if authManager.isAuthenticated {
                MainTabView(
                    authManager: authManager,
                    apiClient: apiClient
                )
            } else {
                AuthView(authManager: authManager)
            }
        }
        .animation(.default, value: authManager.isAuthenticated)
        .task {
            await authManager.restoreSession()
        }
    }
}
