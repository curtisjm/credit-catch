import SwiftUI

struct AuthView: View {
    @State private var viewModel: AuthViewModel

    init(authManager: AuthManager) {
        _viewModel = State(initialValue: AuthViewModel(authManager: authManager))
    }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 24) {
                    // Header
                    VStack(spacing: 8) {
                        Text("Credit Catch")
                            .font(.largeTitle.bold())
                        Text(viewModel.isSignUp ? "Create your account" : "Welcome back")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                    .padding(.top, 40)

                    // Form fields
                    VStack(spacing: 16) {
                        if viewModel.isSignUp {
                            TextField("Display name", text: $viewModel.displayName)
                                .textContentType(.name)
                                .textInputAutocapitalization(.words)
                        }

                        TextField("Email", text: $viewModel.email)
                            .textContentType(.emailAddress)
                            .textInputAutocapitalization(.never)
                            .keyboardType(.emailAddress)

                        SecureField("Password", text: $viewModel.password)
                            .textContentType(viewModel.isSignUp ? .newPassword : .password)
                    }
                    .textFieldStyle(.roundedBorder)

                    // Error
                    if let error = viewModel.errorMessage {
                        Text(error)
                            .font(.footnote)
                            .foregroundStyle(.red)
                            .multilineTextAlignment(.center)
                    }

                    // Submit
                    Button {
                        Task { await viewModel.submit() }
                    } label: {
                        Group {
                            if viewModel.isSubmitting {
                                ProgressView()
                            } else {
                                Text(viewModel.isSignUp ? "Create Account" : "Sign In")
                            }
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 12)
                    }
                    .buttonStyle(.borderedProminent)
                    .disabled(!viewModel.isFormValid || viewModel.isSubmitting)

                    // Toggle mode
                    Button {
                        withAnimation { viewModel.isSignUp.toggle() }
                    } label: {
                        Text(viewModel.isSignUp
                             ? "Already have an account? Sign in"
                             : "Don't have an account? Sign up")
                            .font(.footnote)
                    }
                }
                .padding(.horizontal, 24)
            }
            .navigationBarHidden(true)
        }
    }
}
