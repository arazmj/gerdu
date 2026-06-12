# GerduKit — Gerdu client for iOS and other Apple platforms

`GerduKit` is a small, dependency-free Swift package for talking to a
[Gerdu](../../../README.md) server from an **iOS** app (or macOS, tvOS, and
watchOS). It wraps Gerdu's always-on HTTP API using only Foundation's
`URLSession`, so there is nothing extra to install in your app.

| | |
|---|---|
| Platforms | iOS 13+, macOS 10.15+, tvOS 13+, watchOS 6+ |
| Swift     | 5.7+ |
| Dependencies | none (Foundation only) |

## Why this exists

Gerdu is a server, so "supporting iOS" means giving iOS apps an easy,
idiomatic way to talk to it. Any iOS app can reach a Gerdu node through the
HTTP API:

```text
PUT    /cache/{key}   body = value   -> 201 Created or 204 No Content
GET    /cache/{key}                  -> 200 OK + value, or 404 Not Found
DELETE /cache/{key}                  -> 200 OK, or 404 Not Found
```

`GerduClient` turns those three operations into `put`, `get`, and `delete`
calls with both `async`/`await` and completion-handler flavors.

## Install

### Swift Package Manager (Xcode)

In Xcode choose **File ▸ Add Package Dependencies…** and point it at the Gerdu
repository, then select the **GerduKit** product. The package lives in
`examples/iOS/GerduKit`.

### Swift Package Manager (Package.swift)

```swift
dependencies: [
    .package(url: "https://github.com/arazmj/gerdu.git", branch: "master"),
],
targets: [
    .target(
        name: "MyApp",
        dependencies: [
            .product(name: "GerduKit", package: "gerdu"),
        ]
    ),
]
```

You can also drag the `GerduKit` folder into your project, or depend on it by
local path while experimenting:

```swift
.package(path: "../gerdu/examples/iOS/GerduKit")
```

## Quick start (async/await)

```swift
import GerduKit

let client = try GerduClient(host: "127.0.0.1", port: 8080)

// Store a value (created == the key was new, updated == it already existed).
let outcome = try await client.put(key: "greeting", value: "hello")
print(outcome == .created ? "inserted" : "updated")

// Read it back; nil means the key is not present.
if let value = try await client.get(key: "greeting") {
    print("greeting =", value)
}

// Delete it; returns true if the key existed.
let removed = try await client.delete(key: "greeting")
print("deleted:", removed)
```

A typical SwiftUI view model:

```swift
import SwiftUI
import GerduKit

@MainActor
final class CacheViewModel: ObservableObject {
    @Published var value: String = ""

    private let client = try! GerduClient(host: "cache.example.com", port: 8080, useTLS: true)

    func load(key: String) async {
        do {
            value = try await client.get(key: key) ?? "(missing)"
        } catch {
            value = "error: \(error.localizedDescription)"
        }
    }
}
```

## Quick start (completion handlers)

For deployment targets without `async`/`await`, the same operations are
available with completion handlers:

```swift
client.put(key: "greeting", value: "hello") { result in
    switch result {
    case .success(let outcome): print(outcome)        // .created or .updated
    case .failure(let error):   print(error)
    }
}
```

## App Transport Security (HTTP vs HTTPS)

iOS blocks plain `http://` traffic by default through App Transport Security
(ATS). For production, run Gerdu with TLS (`-cert` / `-key`) and connect with
`useTLS: true`.

For **local development** against an `http://` server, add an ATS exception to
your app's `Info.plist`. Scope it as tightly as possible, for example:

```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSExceptionDomains</key>
    <dict>
        <key>127.0.0.1</key>
        <dict>
            <key>NSExceptionAllowsInsecureHTTPLoads</key>
            <true/>
        </dict>
    </dict>
</dict>
```

## Running a Gerdu server to test against

From the repository root:

```bash
go build -v .
./gerdu -host 0.0.0.0 -httpport 8080
```

Then point `GerduClient` at that host/port.

## API reference

```swift
// Create a client.
init(baseURL: URL, session: URLSession = .shared)
init(host: String, port: Int = 8080, useTLS: Bool = false, session: URLSession = .shared) throws

// Operations (async/await).
func put(key: String, value: String) async throws -> PutResult   // .created or .updated
func get(key: String) async throws -> String?                    // nil when absent
func delete(key: String) async throws -> Bool                    // true when the key existed

// Operations (completion handlers).
func put(key:value:completion:)
func get(key:completion:)
func delete(key:completion:)
```

`PutResult` is `.created` (HTTP 201) or `.updated` (HTTP 204). Errors are
reported as `GerduError`.

## Running the package tests

```bash
cd examples/iOS/GerduKit
swift test
```

The tests stub HTTP responses with a custom `URLProtocol`, so they run without
a live Gerdu server.

## gRPC alternative

If you prefer gRPC over HTTP, Gerdu also ships a gRPC service defined in
[`proto/gerdu.proto`](../../../proto/gerdu.proto). See the existing
[Objective-C gRPC example](../../gRPC/Objective-C) for an Apple-platform gRPC
client built with CocoaPods.
