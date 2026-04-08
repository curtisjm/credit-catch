import SwiftUI

/// Root view that switches between auth and main flows based on authentication state.
struct RootView: View {
    let authManager: AuthManager

    var body: some View {
        Group {
            if authManager.isLoading {
                ProgressView("Loading...")
            } else if authManager.isAuthenticated {
                HomeView(authManager: authManager)
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
