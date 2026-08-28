//
//  GerduClientTests.swift
//  GerduKitTests
//
//  Unit tests for GerduClient. These do not require a running Gerdu server:
//  HTTP responses are stubbed with a custom URLProtocol so the request
//  building and status-code mapping can be verified in isolation.
//

import XCTest
@testable import GerduKit

final class GerduClientTests: XCTestCase {

    // MARK: - URL building

    func testCacheURLBuildsExpectedPath() throws {
        let client = GerduClient(baseURL: URL(string: "http://127.0.0.1:8080")!)
        let url = try client.cacheURL(for: "Hello")
        XCTAssertEqual(url.absoluteString, "http://127.0.0.1:8080/cache/Hello")
    }

    func testCacheURLTrimsTrailingSlashOnBaseURL() throws {
        let client = GerduClient(baseURL: URL(string: "http://127.0.0.1:8080/")!)
        let url = try client.cacheURL(for: "greeting")
        XCTAssertEqual(url.absoluteString, "http://127.0.0.1:8080/cache/greeting")
    }

    func testCacheURLPercentEncodesKey() throws {
        let client = GerduClient(baseURL: URL(string: "http://127.0.0.1:8080")!)
        let url = try client.cacheURL(for: "a b/c")
        // Spaces and slashes must be encoded so the key stays in one path segment.
        XCTAssertEqual(url.absoluteString, "http://127.0.0.1:8080/cache/a%20b%2Fc")
    }

    func testCacheURLRejectsEmptyKey() {
        let client = GerduClient(baseURL: URL(string: "http://127.0.0.1:8080")!)
        XCTAssertThrowsError(try client.cacheURL(for: "")) { error in
            XCTAssertEqual(error as? GerduError, .invalidKey(""))
        }
    }

    func testConvenienceInitBuildsBaseURL() throws {
        let plain = try GerduClient(host: "cache.example.com", port: 9000)
        XCTAssertEqual(plain.baseURL.absoluteString, "http://cache.example.com:9000")

        let secure = try GerduClient(host: "cache.example.com", port: 443, useTLS: true)
        XCTAssertEqual(secure.baseURL.absoluteString, "https://cache.example.com:443")
    }

    // MARK: - PUT

    func testPutReturnsCreatedOn201() async throws {
        let client = makeClient { request in
            XCTAssertEqual(request.httpMethod, "PUT")
            XCTAssertEqual(request.url?.path, "/cache/Hello")
            return .init(statusCode: 201, data: Data())
        }
        let result = try await client.put(key: "Hello", value: "World")
        XCTAssertEqual(result, .created)
    }

    func testPutReturnsUpdatedOn204() async throws {
        let client = makeClient { _ in .init(statusCode: 204, data: Data()) }
        let result = try await client.put(key: "Hello", value: "World")
        XCTAssertEqual(result, .updated)
    }

    func testPutThrowsOnUnexpectedStatus() async {
        let client = makeClient { _ in .init(statusCode: 500, data: Data()) }
        await assertThrowsGerduError(.unexpectedStatus(500)) {
            _ = try await client.put(key: "Hello", value: "World")
        }
    }

    // MARK: - GET

    func testGetReturnsValueOn200() async throws {
        let client = makeClient { request in
            XCTAssertEqual(request.httpMethod, "GET")
            return .init(statusCode: 200, data: Data("World".utf8))
        }
        let value = try await client.get(key: "Hello")
        XCTAssertEqual(value, "World")
    }

    func testGetReturnsNilOn404() async throws {
        let client = makeClient { _ in .init(statusCode: 404, data: Data()) }
        let value = try await client.get(key: "missing")
        XCTAssertNil(value)
    }

    func testGetThrowsOnUnexpectedStatus() async {
        let client = makeClient { _ in .init(statusCode: 503, data: Data()) }
        await assertThrowsGerduError(.unexpectedStatus(503)) {
            _ = try await client.get(key: "Hello")
        }
    }

    // MARK: - DELETE

    func testDeleteReturnsTrueOn200() async throws {
        let client = makeClient { request in
            XCTAssertEqual(request.httpMethod, "DELETE")
            return .init(statusCode: 200, data: Data())
        }
        let deleted = try await client.delete(key: "Hello")
        XCTAssertTrue(deleted)
    }

    func testDeleteReturnsFalseOn404() async throws {
        let client = makeClient { _ in .init(statusCode: 404, data: Data()) }
        let deleted = try await client.delete(key: "missing")
        XCTAssertFalse(deleted)
    }

    // MARK: - Helpers

    private func makeClient(handler: @escaping (URLRequest) throws -> MockURLProtocol.Stub) -> GerduClient {
        let configuration = URLSessionConfiguration.ephemeral
        configuration.protocolClasses = [MockURLProtocol.self]
        let session = URLSession(configuration: configuration)
        MockURLProtocol.requestHandler = handler
        return GerduClient(baseURL: URL(string: "http://127.0.0.1:8080")!, session: session)
    }

    private func assertThrowsGerduError(
        _ expected: GerduError,
        _ body: () async throws -> Void,
        file: StaticString = #filePath,
        line: UInt = #line
    ) async {
        do {
            try await body()
            XCTFail("Expected to throw \(expected)", file: file, line: line)
        } catch let error as GerduError {
            XCTAssertEqual(error, expected, file: file, line: line)
        } catch {
            XCTFail("Threw unexpected error \(error)", file: file, line: line)
        }
    }
}

/// A `URLProtocol` that returns canned HTTP responses so tests run without a server.
final class MockURLProtocol: URLProtocol {

    struct Stub {
        let statusCode: Int
        let data: Data
    }

    static var requestHandler: ((URLRequest) throws -> Stub)?

    override class func canInit(with request: URLRequest) -> Bool { true }

    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        guard let handler = MockURLProtocol.requestHandler else {
            client?.urlProtocol(self, didFailWithError: GerduError.invalidResponse)
            return
        }
        do {
            let stub = try handler(request)
            let response = HTTPURLResponse(
                url: request.url!,
                statusCode: stub.statusCode,
                httpVersion: "HTTP/1.1",
                headerFields: nil
            )!
            client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
            client?.urlProtocol(self, didLoad: stub.data)
            client?.urlProtocolDidFinishLoading(self)
        } catch {
            client?.urlProtocol(self, didFailWithError: error)
        }
    }

    override func stopLoading() {}
}
