import SwiftUI

struct HomeView: View {
    let authManager: AuthManager

    var body: some View {
        NavigationStack {
            VStack(spacing: 20) {
                if let user = authManager.currentUser {
                    Text("Hello, \(user.displayName ?? user.email)")
                        .font(.title2)
                }

                Text("Credit Catch")
                    .font(.headline)
                    .foregroundStyle(.secondary)
            }
            .navigationTitle("Home")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Sign Out") {
                        authManager.signOut()
                    }
                }
            }
        }
    }
}
