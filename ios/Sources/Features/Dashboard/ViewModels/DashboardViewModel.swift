import Foundation
import Observation

@Observable
@MainActor
final class DashboardViewModel {
    private(set) var greeting: String = ""
    private(set) var creditScore: Int?
    private(set) var isLoading = false
    var errorMessage: String?

    private let authManager: AuthManager
    private let apiClient: APIClient

    init(authManager: AuthManager, apiClient: APIClient) {
        self.authManager = authManager
        self.apiClient = apiClient
        updateGreeting()
    }

    var userName: String {
        authManager.currentUser?.displayName ?? authManager.currentUser?.email ?? "there"
    }

    func loadDashboard() async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        // TODO: Fetch dashboard data from API when endpoint is ready
        // do {
        //     let summary: DashboardSummary = try await apiClient.request(.get("/dashboard"))
        //     self.creditScore = summary.creditScore
        // } catch {
        //     errorMessage = error.localizedDescription
        // }
    }

    private func updateGreeting() {
        let hour = Calendar.current.component(.hour, from: Date())
        switch hour {
        case 5..<12: greeting = "Good morning"
        case 12..<17: greeting = "Good afternoon"
        default: greeting = "Good evening"
        }
    }
}
