import Foundation

/// Provides auth tokens to the networking layer.
/// Implemented by AuthManager — keeps APIClient decoupled from auth storage.
protocol AuthTokenProvider: Sendable {
    func accessToken() async -> String?
    func refreshToken() async throws -> String
    func clearAuth() async
}

/// Central networking client. Handles request construction, auth token injection,
/// automatic token refresh on 401, and JSON decoding.
///
/// Usage:
///     let user: User = try await api.request(.get("/me"))
///
actor APIClient {
    private let baseURL: URL
    private let session: URLSession
    private let decoder: JSONDecoder
    private let authProvider: (any AuthTokenProvider)?

    init(
        baseURL: URL,
        session: URLSession = .shared,
        decoder: JSONDecoder = .defaultAPI,
        authProvider: (any AuthTokenProvider)? = nil
    ) {
        self.baseURL = baseURL
        self.session = session
        self.decoder = decoder
        self.authProvider = authProvider
    }

    // MARK: - Public

    /// Perform a request and decode the response as `T`.
    func request<T: Decodable & Sendable>(_ endpoint: APIEndpoint) async throws -> T {
        let data = try await perform(endpoint, retryOnUnauth: true)
        do {
            return try decoder.decode(T.self, from: data)
        } catch {
            throw APIError.decoding(error)
        }
    }

    /// Perform a request that returns no meaningful body (e.g. DELETE 204).
    func requestVoid(_ endpoint: APIEndpoint) async throws {
        _ = try await perform(endpoint, retryOnUnauth: true)
    }

    // MARK: - Internal pipeline

    private func perform(_ endpoint: APIEndpoint, retryOnUnauth: Bool) async throws -> Data {
        guard let url = endpoint.url(baseURL: baseURL) else {
            throw APIError.invalidURL
        }

        var request = URLRequest(url: url)
        request.httpMethod = endpoint.method.rawValue
        request.httpBody = endpoint.body

        // Merge default + endpoint headers
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        for (key, value) in endpoint.headers {
            request.setValue(value, forHTTPHeaderField: key)
        }

        // Inject auth token
        if endpoint.requiresAuth, let provider = authProvider {
            if let token = await provider.accessToken() {
                request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
            }
        }

        let (data, response): (Data, URLResponse)
        do {
            (data, response) = try await session.data(for: request)
        } catch {
            throw APIError.network(error)
        }

        guard let http = response as? HTTPURLResponse else {
            throw APIError.network(URLError(.badServerResponse))
        }

        switch http.statusCode {
        case 200...299:
            return data
        case 401:
            if retryOnUnauth, let provider = authProvider {
                do {
                    _ = try await provider.refreshToken()
                    return try await perform(endpoint, retryOnUnauth: false)
                } catch {
                    await provider.clearAuth()
                    throw APIError.tokenRefreshFailed
                }
            }
            throw APIError.unauthorized
        case 403:
            throw APIError.forbidden
        case 404:
            throw APIError.notFound
        default:
            throw APIError.server(statusCode: http.statusCode, data: data)
        }
    }
}

// MARK: - Default JSON decoder

extension JSONDecoder {
    static let defaultAPI: JSONDecoder = {
        let d = JSONDecoder()
        d.keyDecodingStrategy = .convertFromSnakeCase
        d.dateDecodingStrategy = .iso8601
        return d
    }()
}
