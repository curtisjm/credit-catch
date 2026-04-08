import Foundation

struct TokenPair: Codable, Sendable {
    let accessToken: String
    let refreshToken: String
}
