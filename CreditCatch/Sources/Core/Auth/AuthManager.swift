import Foundation
import Observation

/// Manages authentication state, token persistence, and provides tokens to APIClient.
///
/// Observed by SwiftUI views to drive auth-gated navigation.
/// Implements `AuthTokenProvider` (via a bridge) so the networking layer can
/// inject tokens and trigger refresh without knowing about Keychain.
@Observable
@MainActor
final class AuthManager {
    // MARK: - Observable state

    private(set) var currentUser: User?
    private(set) var isLoading = true

    var isAuthenticated: Bool { currentUser != nil }

    // MARK: - Private

    private let keychain: KeychainService
    private let baseURL: URL
    private let session: URLSession

    private static let accessTokenKey = "access_token"
    private static let refreshTokenKey = "refresh_token"

    init(
        baseURL: URL,
        keychain: KeychainService = KeychainService(),
        session: URLSession = .shared
    ) {
        self.baseURL = baseURL
        self.keychain = keychain
        self.session = session
    }

    // MARK: - Public API

    /// Call at app launch to restore a saved session.
    func restoreSession() async {
        isLoading = true
        defer { isLoading = false }

        guard storedAccessToken() != nil else { return }

        // Validate the token by fetching the user profile
        do {
            let api = makeAPIClient()
            let user: User = try await api.request(.get("/auth/me"))
            self.currentUser = user
        } catch {
            // Token invalid — clear and require fresh login
            clearTokens()
        }
    }

    /// Sign in with email/password.
    func signIn(email: String, password: String) async throws {
        struct Credentials: Encodable {
            let email: String
            let password: String
        }

        guard let endpoint = APIEndpoint.post(
            "/auth/login",
            body: Credentials(email: email, password: password),
            requiresAuth: false
        ) else {
            throw APIError.invalidURL
        }

        let api = makeAPIClient()

        struct LoginResponse: Decodable {
            let tokens: TokenPair
            let user: User
        }

        let response: LoginResponse = try await api.request(endpoint)
        try storeTokens(response.tokens)
        self.currentUser = response.user
    }

    /// Register a new account.
    func signUp(email: String, password: String, displayName: String?) async throws {
        struct Registration: Encodable {
            let email: String
            let password: String
            let displayName: String?
        }

        guard let endpoint = APIEndpoint.post(
            "/auth/register",
            body: Registration(email: email, password: password, displayName: displayName),
            requiresAuth: false
        ) else {
            throw APIError.invalidURL
        }

        let api = makeAPIClient()

        struct RegisterResponse: Decodable {
            let tokens: TokenPair
            let user: User
        }

        let response: RegisterResponse = try await api.request(endpoint)
        try storeTokens(response.tokens)
        self.currentUser = response.user
    }

    /// Sign out — clears tokens and resets state.
    func signOut() {
        clearTokens()
        currentUser = nil
    }

    // MARK: - APIClient factory

    /// Creates an APIClient wired to this AuthManager for token injection.
    func makeAPIClient() -> APIClient {
        APIClient(
            baseURL: baseURL,
            session: session,
            authProvider: AuthBridge(manager: self)
        )
    }

    // MARK: - Token storage

    fileprivate func storeTokens(_ pair: TokenPair) throws {
        try keychain.save(Data(pair.accessToken.utf8), for: Self.accessTokenKey)
        try keychain.save(Data(pair.refreshToken.utf8), for: Self.refreshTokenKey)
    }

    fileprivate func storedAccessToken() -> String? {
        guard let data = try? keychain.load(Self.accessTokenKey) else { return nil }
        return String(data: data, encoding: .utf8)
    }

    fileprivate func storedRefreshToken() -> String? {
        guard let data = try? keychain.load(Self.refreshTokenKey) else { return nil }
        return String(data: data, encoding: .utf8)
    }

    fileprivate func clearTokens() {
        try? keychain.delete(Self.accessTokenKey)
        try? keychain.delete(Self.refreshTokenKey)
    }

    /// Attempts to refresh the access token using the stored refresh token.
    fileprivate func refreshAccessToken() async throws -> String {
        guard let refreshToken = storedRefreshToken() else {
            throw APIError.tokenRefreshFailed
        }

        struct RefreshRequest: Encodable {
            let refreshToken: String
        }

        struct RefreshResponse: Decodable {
            let tokens: TokenPair
        }

        var request = URLRequest(url: baseURL.appendingPathComponent("/auth/refresh"))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(RefreshRequest(refreshToken: refreshToken))

        let (data, response) = try await session.data(for: request)

        guard let http = response as? HTTPURLResponse, http.statusCode == 200 else {
            throw APIError.tokenRefreshFailed
        }

        let decoded = try JSONDecoder.defaultAPI.decode(RefreshResponse.self, from: data)
        try storeTokens(decoded.tokens)
        return decoded.tokens.accessToken
    }
}

// MARK: - Bridge to AuthTokenProvider

/// Bridges MainActor-isolated AuthManager to the Sendable AuthTokenProvider protocol.
/// The APIClient (an actor) calls these methods; Swift handles the actor hop automatically.
private final class AuthBridge: AuthTokenProvider, @unchecked Sendable {
    let manager: AuthManager

    init(manager: AuthManager) {
        self.manager = manager
    }

    func accessToken() async -> String? {
        await manager.storedAccessToken()
    }

    func refreshToken() async throws -> String {
        try await manager.refreshAccessToken()
    }

    func clearAuth() async {
        await manager.signOut()
    }
}
