use std::time::{SystemTime, UNIX_EPOCH};
use serde::{Serialize, Deserialize};
use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use dotenv::dotenv;

const PORT: &str = "3000";

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Test {
    x: f64,
    y: f64,
}

#[derive(Debug, Serialize, Deserialize)]
struct TestRequest {
    secret: String,
    tests: Vec<Test>,
    budget: f64,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct TestResponse {
    test_results: Vec<Test>,
    server_start: f64,
    server_end: f64,
}

async fn process_test_data(tests: &mut Vec<Test>, budget: f64) -> Vec<Test> {
    tests.sort_unstable_by(|a, b| {
        (a.x / a.y)
            .partial_cmp(&(b.x / b.y))
            .unwrap_or(std::cmp::Ordering::Equal)
    });
    let mut spent = 0.0;
    let mut results = Vec::new();
    for test in tests {
        if spent + test.y <= budget {
            results.push(test.clone());
            spent += test.y;
        }
    }
    results
}

async fn run_test_handler(test_request: web::Json<TestRequest>) -> impl Responder {
    if test_request.secret != std::env::var("API_KEY").unwrap_or_default() {
        println!("request secret: {}, local secret: {}", test_request.secret, std::env::var("API_KEY").unwrap_or_default());
        return HttpResponse::Forbidden().finish();
    }

    let start_time = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs_f64();
    let mut tests = test_request.tests.clone();
    let budget = test_request.budget.clone();

    let resp = process_test_data(&mut tests, budget).await;

    let end_time = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs_f64();

    HttpResponse::Ok().json(TestResponse {
        test_results: resp,
        server_start: start_time,
        server_end: end_time,
    })
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenv().ok();
    let address = format!("0.0.0.0:{}", PORT);
    HttpServer::new(|| App::new().service(web::resource("/runtest").route(web::post().to(run_test_handler))))
        .bind(&address)?
        .run()
        .await
}