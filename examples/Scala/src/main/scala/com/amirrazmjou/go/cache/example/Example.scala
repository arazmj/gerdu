package com.amirrazmjou.go.cache.example

import scalaj.http._

object Example extends App {
  private val url = "http://localhost:8080/cache/Hello"

  Http(url)
    .postData("World")
    .method("put")
    .asString

  val value = Http(url).asString

  print("Hello = " + value.body)
}
