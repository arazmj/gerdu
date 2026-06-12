//
//  GerduClient.swift
//  GerduKit
//
//  A small, dependency-free client for talking to a Gerdu server from
//  Apple platforms (iOS, macOS, tvOS, watchOS).
//
//  Gerdu always exposes an HTTP API for key/value operations:
//
//      PUT    /cache/{key}   body = value   -> 201 Created or 204 No Content
//      GET    /cache/{key}                  -> 200 OK + value, or 404 Not Found
//      DELETE /cache/{key}                  -> 200 OK, or 404 Not Found
//
//  This file wraps those three operations using only Foundation's
//  URLSession, which is fully available on iOS, so it can be dropped into
//  any iOS app without extra packages.
//

import Foundation
#if canImport(FoundationNetworking)
import FoundationNetworking
#endif

/// Errors thrown by ``GerduClient``.
public enum GerduError: Error, Equatable {
    /// The supplied key was empty or could not be percent-encoded.
    case invalidKey(String)
    /// The host/port could not be turned into a valid base URL.
    case invalidBaseURL(String)
    /// The server response was not an HTTP response.
    case invalidResponse
    /// The server returned a status code the client does not expect for the operation.
    case unexpectedStatus(Int)
}

extension GerduError: LocalizedError {
    public var errorDescription: String? {
        switch self {
        case .invalidKey(let key):
            return "Invalid Gerdu key: \"\(key)\""
        case .invalidBaseURL(let url):
            return "Invalid Gerdu base URL: \"\(url)\""
        case .invalidResponse:
            return "The server did not return a valid HTTP response."
        case .unexpectedStatus(let code):
            return "The server returned an unexpected HTTP status code: \(code)"
        }
    }
}

/// The outcome of a successful ``GerduClient/put(key:value:completion:)`` call.
public enum PutResult: Equatable {
    /// The key did not exist before and was created (HTTP 201).
    case created
    /// An existing key was overwritten (HTTP 204).
    case updated
}

/// A client for a single Gerdu server endpoint.
///
/// `GerduClient` is a lightweight value type that you can freely copy and
/// share. All operations are asynchronous and non-blocking, so they are safe
/// to call from the main thread of an iOS app.
public struct GerduClient {

    /// The base URL of the Gerdu server, e.g. `http://127.0.0.1:8080`.
    public let baseURL: URL

    private let session: URLSession

    /// Creates a client that targets the given base URL.
    ///
    /// - Parameters:
    ///   - baseURL: The root URL of the Gerdu HTTP server, without a trailing
    ///     `/cache` path, for example `URL(string: "http://127.0.0.1:8080")!`.
    ///   - session: The `URLSession` used for requests. Defaults to `.shared`.
    public init(baseURL: URL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
    }

    /// Convenience initializer that builds the base URL from a host and port.
    ///
    /// - Parameters:
    ///   - host: The server host, e.g. `"127.0.0.1"` or `"cache.example.com"`.
    ///   - port: The HTTP port Gerdu listens on. Defaults to `8080`.
    ///   - useTLS: Pass `true` to use `https` (Gerdu's `-cert`/`-key` flags).
    ///   - session: The `URLSession` used for requests. Defaults to `.shared`.
    /// - Throws: ``GerduError/invalidBaseURL(_:)`` if the URL cannot be formed.
    public init(host: String, port: Int = 8080, useTLS: Bool = false, session: URLSession = .shared) throws {
        let scheme = useTLS ? "https" : "http"
        let string = "\(scheme)://\(host):\(port)"
        guard let url = URL(string: string) else {
            throw GerduError.invalidBaseURL(string)
        }
        self.init(baseURL: url, session: session)
    }

    // MARK: - Completion-handler API

    /// Stores `value` under `key`, creating it if necessary.
    public func put(key: String, value: String, completion: @escaping (Result<PutResult, Error>) -> Void) {
        let request: URLRequest
        do {
            var r = URLRequest(url: try cacheURL(for: key))
            r.httpMethod = "PUT"
            r.httpBody = Data(value.utf8)
            request = r
        } catch {
            completion(.failure(error))
            return
        }
        perform(request) { result in
            completion(result.flatMap { status, _ in
                switch status {
                case 201: return .success(.created)
                case 204: return .success(.updated)
                default: return .failure(GerduError.unexpectedStatus(status))
                }
            })
        }
    }

    /// Retrieves the value stored under `key`, or `nil` if the key is absent.
    public func get(key: String, completion: @escaping (Result<String?, Error>) -> Void) {
        let request: URLRequest
        do {
            var r = URLRequest(url: try cacheURL(for: key))
            r.httpMethod = "GET"
            request = r
        } catch {
            completion(.failure(error))
            return
        }
        perform(request) { result in
            completion(result.flatMap { status, data in
                switch status {
                case 200:
                    let value: String? = String(decoding: data, as: UTF8.self)
                    return .success(value)
                case 404:
                    return .success(nil)
                default:
                    return .failure(GerduError.unexpectedStatus(status))
                }
            })
        }
    }

    /// Deletes `key`. The result is `true` if the key existed, `false` otherwise.
    public func delete(key: String, completion: @escaping (Result<Bool, Error>) -> Void) {
        let request: URLRequest
        do {
            var r = URLRequest(url: try cacheURL(for: key))
            r.httpMethod = "DELETE"
            request = r
        } catch {
            completion(.failure(error))
            return
        }
        perform(request) { result in
            completion(result.flatMap { status, _ in
                switch status {
                case 200: return .success(true)
                case 404: return .success(false)
                default: return .failure(GerduError.unexpectedStatus(status))
                }
            })
        }
    }

    // MARK: - async/await API

    /// Stores `value` under `key`, creating it if necessary.
    @available(iOS 13.0, macOS 10.15, tvOS 13.0, watchOS 6.0, *)
    public func put(key: String, value: String) async throws -> PutResult {
        try await withCheckedThrowingContinuation { continuation in
            put(key: key, value: value) { continuation.resume(with: $0) }
        }
    }

    /// Retrieves the value stored under `key`, or `nil` if the key is absent.
    @available(iOS 13.0, macOS 10.15, tvOS 13.0, watchOS 6.0, *)
    public func get(key: String) async throws -> String? {
        try await withCheckedThrowingContinuation { continuation in
            get(key: key) { continuation.resume(with: $0) }
        }
    }

    /// Deletes `key`. The result is `true` if the key existed, `false` otherwise.
    @available(iOS 13.0, macOS 10.15, tvOS 13.0, watchOS 6.0, *)
    public func delete(key: String) async throws -> Bool {
        try await withCheckedThrowingContinuation { continuation in
            delete(key: key) { continuation.resume(with: $0) }
        }
    }

    // MARK: - Internals

    /// Builds the `/cache/{key}` URL, percent-encoding the key so that it
    /// stays within a single path segment (Gerdu routes keys as one segment).
    func cacheURL(for key: String) throws -> URL {
        guard !key.isEmpty,
              let encoded = key.addingPercentEncoding(withAllowedCharacters: GerduClient.keyAllowedCharacters)
        else {
            throw GerduError.invalidKey(key)
        }
        var root = baseURL.absoluteString
        if root.hasSuffix("/") {
            root.removeLast()
        }
        guard let url = URL(string: "\(root)/cache/\(encoded)") else {
            throw GerduError.invalidKey(key)
        }
        return url
    }

    private func perform(_ request: URLRequest, completion: @escaping (Result<(Int, Data), Error>) -> Void) {
        let task = session.dataTask(with: request) { data, response, error in
            if let error = error {
                completion(.failure(error))
                return
            }
            guard let http = response as? HTTPURLResponse else {
                completion(.failure(GerduError.invalidResponse))
                return
            }
            completion(.success((http.statusCode, data ?? Data())))
        }
        task.resume()
    }

    /// URL path characters allowed in a key. `/` is removed so that keys
    /// containing slashes are encoded instead of splitting the path.
    private static let keyAllowedCharacters: CharacterSet = {
        var set = CharacterSet.urlPathAllowed
        set.remove(charactersIn: "/")
        return set
    }()
}
