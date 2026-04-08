import Foundation
import Observation

@Observable
@MainActor
final class AuthViewModel {
    var email = ""
    var password = ""
    var displayName = ""
    var isSignUp = false
    var isSubmitting = false
    var errorMessage: String?

    private let authManager: AuthManager

    init(authManager: AuthManager) {
        self.authManager = authManager
    }

    var isFormValid: Bool {
        !email.isEmpty && password.count >= 8
    }

    func submit() async {
        guard isFormValid else { return }
        isSubmitting = true
        errorMessage = nil

        do {
            if isSignUp {
                try await authManager.signUp(
                    email: email,
                    password: password,
                    displayName: displayName.isEmpty ? nil : displayName
                )
            } else {
                try await authManager.signIn(email: email, password: password)
            }
        } catch {
            errorMessage = error.localizedDescription
        }

        isSubmitting = false
    }
}
