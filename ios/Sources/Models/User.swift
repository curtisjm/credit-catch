import Foundation

struct User: Codable, Sendable, Identifiable {
    let id: String
    let email: String
    let displayName: String?
    let createdAt: Date?
}
