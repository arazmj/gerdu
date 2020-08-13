
#[tokio::main]
async fn main()  -> Result<(), Box<dyn std::error::Error>> {
    let url = "http://localhost:8080/cache/Hello";

    let client = reqwest::Client::new();
    client.post(url)
        .body("World")
        .send()
        .await?;

    let body = reqwest::get(url)
        .await?
        .text()
        .await?;

    println!("Hello = {:?}", body);
    Ok(())
}
