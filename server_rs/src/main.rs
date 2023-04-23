use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use dotenv::dotenv;
use serde::{Deserialize, Serialize};
use std::time::SystemTime;

const PORT: &str = "3000";

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Test {
    value: i64,
}

#[derive(Debug, Serialize, Deserialize)]
struct TestRequest {
    secret: String,
    tests: Vec<Test>,
    budget: i64,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct TestResponse {
    test_results: Vec<Test>,
    duration: f64,
}

async fn process_test_data(mut tests: Vec<Test>, budget: i64) -> Vec<Test> {
    tests.sort_unstable_by_key(|test| test.value);
    tests
        .into_iter()
        .take_while(|test| test.value < budget)
        .collect()
}

async fn run_test_handler(test_request: web::Json<TestRequest>) -> impl Responder {
    if test_request.secret != std::env::var("API_KEY").unwrap_or_default() {
        return HttpResponse::Forbidden().finish();
    }

    let tests = test_request.tests.clone();
    let budget = test_request.budget.clone();

    let start_time = SystemTime::now();
    let resp = process_test_data(tests, budget).await;
    let duration = SystemTime::now()
        .duration_since(start_time)
        .unwrap_or_default()
        .as_secs_f64();

    HttpResponse::Ok().json(TestResponse {
        test_results: resp,
        duration: duration,
    })
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenv().ok();
    let address = format!("0.0.0.0:{}", PORT);
    HttpServer::new(|| {
        App::new().service(web::resource("/runtest").route(web::post().to(run_test_handler)))
    })
    .bind(&address)?
    .run()
    .await
}
