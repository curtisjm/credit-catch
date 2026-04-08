import Foundation

/// Centralized configuration and dependency wiring.
/// Future beads add services here — every feature pulls from this single source.
@MainActor
struct AppEnvironment {
    let authManager: AuthManager
    let apiClient: APIClient

    static let shared = AppEnvironment()

    private init() {
        // TODO: Replace with production URL from build config
        let baseURL = URL(string: "https://api.creditcatch.com/v1")!

        let authManager = AuthManager(baseURL: baseURL)
        self.authManager = authManager
        self.apiClient = authManager.makeAPIClient()
    }
}
