import SwiftUI

struct CardsView: View {
    @State private var viewModel: CardsViewModel

    init(apiClient: APIClient) {
        _viewModel = State(wrappedValue: CardsViewModel(apiClient: apiClient))
    }

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading && viewModel.cards.isEmpty {
                    ProgressView()
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                } else if viewModel.cards.isEmpty {
                    ContentUnavailableView(
                        "No Cards Yet",
                        systemImage: "creditcard",
                        description: Text("Your tracked credit cards will appear here.")
                    )
                } else {
                    List(viewModel.cards) { card in
                        CardRow(card: card)
                    }
                }
            }
            .navigationTitle("Cards")
            .refreshable {
                await viewModel.loadCards()
            }
            .task {
                await viewModel.loadCards()
            }
        }
    }
}

// MARK: - Supporting Views

private struct CardRow: View {
    let card: CardSummary

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text(card.name)
                    .font(.headline)
                Text(card.issuer)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }
            Spacer()
            if let lastFour = card.lastFour {
                Text("••••\u{2009}\(lastFour)")
                    .font(.caption.monospaced())
                    .foregroundStyle(.secondary)
            }
        }
        .padding(.vertical, 4)
    }
}
