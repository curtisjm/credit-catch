import Foundation

struct APIEndpoint {
    let path: String
    let method: HTTPMethod
    let headers: [String: String]
    let queryItems: [URLQueryItem]
    let body: Data?
    let requiresAuth: Bool

    init(
        path: String,
        method: HTTPMethod = .get,
        headers: [String: String] = [:],
        queryItems: [URLQueryItem] = [],
        body: Data? = nil,
        requiresAuth: Bool = true
    ) {
        self.path = path
        self.method = method
        self.headers = headers
        self.queryItems = queryItems
        self.body = body
        self.requiresAuth = requiresAuth
    }

    func url(baseURL: URL) -> URL? {
        var components = URLComponents(url: baseURL.appendingPathComponent(path), resolvingAgainstBaseURL: true)
        if !queryItems.isEmpty {
            components?.queryItems = queryItems
        }
        return components?.url
    }
}

// MARK: - Convenience builders

extension APIEndpoint {
    static func get(_ path: String, query: [URLQueryItem] = [], requiresAuth: Bool = true) -> APIEndpoint {
        APIEndpoint(path: path, queryItems: query, requiresAuth: requiresAuth)
    }

    static func post<T: Encodable>(_ path: String, body: T, requiresAuth: Bool = true) -> APIEndpoint? {
        guard let data = try? JSONEncoder().encode(body) else { return nil }
        return APIEndpoint(
            path: path,
            method: .post,
            headers: ["Content-Type": "application/json"],
            body: data,
            requiresAuth: requiresAuth
        )
    }
}
