import SwiftUI

struct SettingsView: View {
    @State private var viewModel: SettingsViewModel

    init(authManager: AuthManager) {
        _viewModel = State(wrappedValue: SettingsViewModel(authManager: authManager))
    }

    var body: some View {
        NavigationStack {
            List {
                // Account section
                Section("Account") {
                    LabeledContent("Email", value: viewModel.userEmail)

                    if !viewModel.userDisplayName.isEmpty {
                        LabeledContent("Name", value: viewModel.userDisplayName)
                    }

                    if !viewModel.memberSince.isEmpty {
                        LabeledContent("Member since", value: viewModel.memberSince)
                    }
                }

                // Preferences section placeholder
                Section("Preferences") {
                    Label("Notifications", systemImage: "bell")
                        .foregroundStyle(.secondary)
                    Label("Appearance", systemImage: "paintbrush")
                        .foregroundStyle(.secondary)
                }

                // Sign out
                Section {
                    Button(role: .destructive) {
                        viewModel.signOut()
                    } label: {
                        HStack {
                            Spacer()
                            Text("Sign Out")
                            Spacer()
                        }
                    }
                }
            }
            .navigationTitle("Settings")
        }
    }
}
