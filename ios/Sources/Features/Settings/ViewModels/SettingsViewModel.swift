import Foundation
import Observation

@Observable
@MainActor
final class SettingsViewModel {
    private let authManager: AuthManager

    init(authManager: AuthManager) {
        self.authManager = authManager
    }

    var userEmail: String {
        authManager.currentUser?.email ?? ""
    }

    var userDisplayName: String {
        authManager.currentUser?.displayName ?? ""
    }

    var memberSince: String {
        guard let date = authManager.currentUser?.createdAt else { return "" }
        return date.formatted(date: .abbreviated, time: .omitted)
    }

    func signOut() {
        authManager.signOut()
    }
}
