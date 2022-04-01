use chrono::{DateTime, Duration, Utc};
use s3::bucket::Bucket;
use s3::creds::Credentials;
use s3::region:: Region;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let (postgres_client, connection): (tokio_postgres::Client, _) = tokio_postgres::connect("postgres://geocloud:geocloud@localhost:5432?sslmode=disable", tokio_postgres::NoTls).await?;

    tokio::spawn(async move {
        if let Err(e) = connection.await {
            panic!("connection error: {}", e);
        }
    });

    let bucket: Bucket = Bucket::new_with_path_style(
        "geocloud", 
        Region::Custom {
            region: "".to_owned(),
            endpoint: "http://127.0.0.1:9000".to_owned()
        }, 
        Credentials::from_profile(Some("default"))?
    )?;

    let delete_before: DateTime<Utc> = Utc::now() - Duration::minutes(1); // Duration::days(14);
    println!("Cleaning up jobs before: {:?}", delete_before);
    for row in postgres_client.query("SELECT * FROM job", &[]).await? {
        let id: &str = row.try_get("job_id")?;
        let end_time: DateTime<Utc> = row.try_get("end_time")?;

        if delete_before > end_time {
            println!("Cleaning up Job ID: {}", id);
            let mut key = "jobs/".to_owned();
            key.push_str(id);
            let results = bucket.list(key.clone(), None).await?;
            for result in results {
                for item in result.contents {
                    bucket.delete_object(item.key).await?;
                }
            }

            bucket.delete_object(key).await?;

            postgres_client.execute("DELETE FROM job WHERE job_id = $1", &[&id]).await?;
        }
    }

    Ok(())
}