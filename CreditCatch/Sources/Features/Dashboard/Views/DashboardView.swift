import SwiftUI

struct DashboardView: View {
    @State private var viewModel: DashboardViewModel

    init(authManager: AuthManager, apiClient: APIClient) {
        _viewModel = State(wrappedValue: DashboardViewModel(
            authManager: authManager,
            apiClient: apiClient
        ))
    }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 24) {
                    // Greeting header
                    VStack(alignment: .leading, spacing: 4) {
                        Text("\(viewModel.greeting),")
                            .font(.title2)
                            .foregroundStyle(.secondary)
                        Text(viewModel.userName)
                            .font(.largeTitle.bold())
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.horizontal)

                    // Credit score card placeholder
                    VStack(spacing: 12) {
                        Text("Credit Score")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)

                        if let score = viewModel.creditScore {
                            Text("\(score)")
                                .font(.system(size: 64, weight: .bold, design: .rounded))
                        } else {
                            Text("—")
                                .font(.system(size: 64, weight: .bold, design: .rounded))
                                .foregroundStyle(.tertiary)
                        }

                        Text("Check back soon")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 32)
                    .background(.regularMaterial, in: RoundedRectangle(cornerRadius: 16))
                    .padding(.horizontal)

                    // Quick actions placeholder
                    VStack(alignment: .leading, spacing: 12) {
                        Text("Quick Actions")
                            .font(.headline)
                            .padding(.horizontal)

                        LazyVGrid(columns: [
                            GridItem(.flexible()),
                            GridItem(.flexible()),
                        ], spacing: 12) {
                            QuickActionCard(icon: "creditcard", title: "My Cards")
                            QuickActionCard(icon: "chart.line.uptrend.xyaxis", title: "Score History")
                            QuickActionCard(icon: "bell", title: "Alerts")
                            QuickActionCard(icon: "magnifyingglass", title: "Compare")
                        }
                        .padding(.horizontal)
                    }

                    if let error = viewModel.errorMessage {
                        Text(error)
                            .font(.caption)
                            .foregroundStyle(.red)
                            .padding(.horizontal)
                    }
                }
                .padding(.vertical)
            }
            .navigationTitle("Dashboard")
            .refreshable {
                await viewModel.loadDashboard()
            }
            .task {
                await viewModel.loadDashboard()
            }
        }
    }
}

// MARK: - Supporting Views

private struct QuickActionCard: View {
    let icon: String
    let title: String

    var body: some View {
        VStack(spacing: 8) {
            Image(systemName: icon)
                .font(.title2)
                .foregroundStyle(.accent)
            Text(title)
                .font(.caption)
                .foregroundStyle(.primary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 16)
        .background(.regularMaterial, in: RoundedRectangle(cornerRadius: 12))
    }
}
