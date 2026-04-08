import Foundation
import Observation

@Observable
@MainActor
final class CardsViewModel {
    private(set) var cards: [CardSummary] = []
    private(set) var isLoading = false
    var errorMessage: String?

    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func loadCards() async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        // TODO: Fetch cards from API when endpoint is ready
        // do {
        //     cards = try await apiClient.request(.get("/cards"))
        // } catch {
        //     errorMessage = error.localizedDescription
        // }
    }
}

/// Lightweight card model for list display. Full model lives in Models/.
struct CardSummary: Identifiable, Sendable {
    let id: String
    let name: String
    let issuer: String
    let lastFour: String?
}
