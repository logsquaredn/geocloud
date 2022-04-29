use chrono::{DateTime, Duration, Utc};
use s3::bucket::Bucket;
use s3::creds::Credentials;
use s3::region:: Region;
use std::io::{Cursor, Write, Seek, SeekFrom};
use std::time::SystemTime;
use clap::Parser;

#[derive(Parser, Debug)]
#[clap(version)]
struct Args {

    #[clap(short, long, default_value_t = 1440)]
    delete_before: i64,

    #[clap(short, long, default_value = "localhost:5432")]
    addr: String,

    #[clap(short, long, default_value = "geocloud")]
    user: String,

    #[clap(short, long)]
    password: String,

    #[clap(short, long, default_value = "disable")]
    sslmode: String,

    #[clap(short = 'r', long, default_value = "geocloud-archive")]
    archive_bucket: String,

    #[clap(short, long, default_value = "geocloud")]
    minio_bucket: String,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();

    let user = args.user;
    let password = args.password;
    let addr = args.addr;
    let sslmode = args.sslmode;
    let conn_string = format!("postgres://{user}:{password}@{addr}?sslmode={sslmode}");
    let (postgres_client, connection): (tokio_postgres::Client, _) = tokio_postgres::connect(&conn_string, tokio_postgres::NoTls).await?;

    tokio::spawn(async move {
        if let Err(e) = connection.await {
            panic!("connection error: {}", e);
        }
    });

    let minio_bucket: Bucket = Bucket::new_with_path_style(
        &args.minio_bucket, 
        Region::Custom {
            region: "".to_owned(),
            endpoint: "http://127.0.0.1:9000".to_owned()
        }, 
        Credentials::from_profile(Some("default"))?
    )?;

    let archive_bucket: Bucket = Bucket::new(
        &args.archive_bucket, 
        "us-east-1".parse()?, 
        Credentials::from_profile(Some("default"))?
    )?;

    let mut archive = Cursor::new(Vec::new());
    archive.write_all("job_id,task_type,job_status,job_error,start_time,end_time,job_args\n".as_bytes())?;
   
    let delete_before: DateTime<Utc> = Utc::now() - Duration::minutes(args.delete_before); // Duration::days(14);
    println!("Cleaning up jobs before: {:?}", delete_before);
    for row in postgres_client.query("SELECT * FROM job", &[]).await? {
        let id: &str = row.try_get("job_id")?;
        let end_time: DateTime<Utc> = row.try_get("end_time")?;

        if delete_before > end_time {
            println!("Cleaning up Job ID: {}", id);
            let mut key = "jobs/".to_owned();
            key.push_str(id);
            key.push_str("/");
            let results = minio_bucket.list(key.clone(), None).await?;
            for result in results {
                for item in result.contents {
                    minio_bucket.delete_object(item.key).await?;
                }
            }

            minio_bucket.delete_object(key).await?;

            postgres_client.execute("DELETE FROM job WHERE job_id = $1", &[&id]).await?;

            dump_row_to_archive(row, &mut archive)?;
        }
    }

    archive.seek(SeekFrom::Start(0))?;
    let mut key = SystemTime::now().duration_since(SystemTime::UNIX_EPOCH)?.as_secs().to_string();
    key.push_str("/archive.csv");
    archive_bucket.put_object_stream(& mut archive, key).await?;

    Ok(())
}

fn dump_row_to_archive(row: tokio_postgres::Row, archive: &mut Cursor<Vec<u8>>) -> Result<(), Box<dyn std::error::Error>> {
    let mut csv_row = "".to_owned();
    let job_id: &str = row.try_get("job_id")?;
    let task_type: &str = row.try_get("task_type")?;
    let job_status: &str = row.try_get("job_status")?;
    let job_error: &str = row.try_get("job_error")?;
    let start_time: DateTime<Utc> = row.try_get("start_time")?;
    let end_time: DateTime<Utc> = row.try_get("end_time")?;
    let job_args: Vec<&str> = row.try_get("job_args")?;

    csv_row.push_str(job_id);
    csv_row.push_str(",");
    csv_row.push_str(task_type);
    csv_row.push_str(",");
    csv_row.push_str(job_status);
    csv_row.push_str(",");
    csv_row.push_str(job_error);
    csv_row.push_str(",");
    csv_row.push_str(&start_time.to_rfc3339());
    csv_row.push_str(",");
    csv_row.push_str(&end_time.to_rfc3339());
    csv_row.push_str(",");
    for arg in job_args {
        csv_row.push_str(arg);
        csv_row.push_str("|");
    }
    csv_row.push_str("\n");

    println!("{}", csv_row);

    archive.write_all(csv_row.as_bytes())?;

    Ok(())
}