import SwiftUI

@main
struct CreditCatchApp: App {
    private let env = AppEnvironment.shared

    var body: some Scene {
        WindowGroup {
            RootView(authManager: env.authManager)
        }
    }
}
