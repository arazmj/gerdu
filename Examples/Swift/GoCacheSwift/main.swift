//
//  main.swift
//  GoCacheSwift
//
//  Created by Amir Razmjou on 8/13/20.
//  Copyright © 2020 Golden Katze. All rights reserved.
//

import Foundation

let dispatchGroup = DispatchGroup()

let baseURL = URL(string: "http://localhost:8080/cache/Hello")!
var request = URLRequest(url: baseURL)
request.httpMethod = "PUT"
let data = try! "World".data(using: String.Encoding.utf8)
URLSession.shared.uploadTask(with: request, from: data) { _, _, _ in
    dispatchGroup.enter()
    URLSession.shared.dataTask(with: baseURL) {(data, response, error) in
        dispatchGroup.leave()
        guard let data = data else { return }
        print("Hello =", String(data: data, encoding: .utf8)!)
    }.resume()
}.resume()

dispatchMain()

