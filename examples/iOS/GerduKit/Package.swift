// swift-tools-version:5.7
//
//  Package.swift
//  GerduKit
//
//  A lightweight, reusable Gerdu client for Apple platforms (iOS, macOS,
//  tvOS, watchOS). Talks to a Gerdu server over its always-on HTTP API
//  using only Foundation's URLSession, so it has no third-party dependencies.
//

import PackageDescription

let package = Package(
    name: "GerduKit",
    platforms: [
        .iOS(.v13),
        .macOS(.v10_15),
        .tvOS(.v13),
        .watchOS(.v6),
    ],
    products: [
        .library(
            name: "GerduKit",
            targets: ["GerduKit"]
        ),
    ],
    targets: [
        .target(
            name: "GerduKit"
        ),
        .testTarget(
            name: "GerduKitTests",
            dependencies: ["GerduKit"]
        ),
    ]
)
