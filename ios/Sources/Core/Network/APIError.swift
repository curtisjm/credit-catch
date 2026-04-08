import Foundation

enum APIError: LocalizedError {
    case invalidURL
    case unauthorized
    case forbidden
    case notFound
    case server(statusCode: Int, data: Data?)
    case decoding(Error)
    case network(Error)
    case tokenRefreshFailed

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            "Invalid URL"
        case .unauthorized:
            "Session expired. Please sign in again."
        case .forbidden:
            "You don't have permission to do that."
        case .notFound:
            "Resource not found."
        case .server(let code, _):
            "Server error (\(code)). Please try again."
        case .decoding:
            "Couldn't read the server response."
        case .network:
            "Network unavailable. Check your connection."
        case .tokenRefreshFailed:
            "Session expired. Please sign in again."
        }
    }
}
