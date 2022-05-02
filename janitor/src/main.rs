use chrono::{DateTime, Duration, Utc};
use s3::bucket::Bucket;
use s3::creds::Credentials;
use s3::region:: Region;
use std::io::{Cursor, Write, Seek, SeekFrom};
use std::time::SystemTime;
use std::env;
use clap::{arg, Command};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Command::new("Janitor")
        .version("0.1.0")
        .arg(arg!(-d --deletebefore <VALUE>).required(false))
        .arg(arg!(-u --postgresuser <VALUE>).required(false))
        .arg(arg!(-p --postgrespassword <VALUE>).required(false))
        .arg(arg!(-a --postgresaddr <VALUE>).required(false))
        .arg(arg!(-s --postgressslmode <VALUE>).required(false))
        .arg(arg!(-r --s3archivebucket <VALUE>).required(false))
        .arg(arg!(-b --s3bucket <VALUE>).required(false))
        .get_matches();

    let postgres_user = match args.value_of("postgresuser") {
        Some(cli_user) => String::from(cli_user),
        None => match env::var("GEOCLOUD_POSTGRES_USER") {
            Ok(env_user) => env_user,
            Err(_e) => String::from("geocloud")
        }
    };

    let postgres_password = match args.value_of("postgrespassword") {
        Some(cli_password) => String::from(cli_password),
        None => match env::var("GEOCLOUD_POSTGRES_PASSWORD") {
            Ok(env_password) => env_password,
            Err(_e) => String::from("")
        }
    };
   
    let postgres_address = match args.value_of("postgresaddr") {
        Some(cli_address) => String::from(cli_address),
        None => match env::var("GEOCLOUD_POSTGRES_ADDRESS") {
            Ok(env_address) => env_address,
            Err(_e) => String::from(":5432")
        }
    };

    let postgres_sslmode = match args.value_of("postgressslmode") {
        Some(cli_sslmode) => cli_sslmode,
        None => "disable"
    };

    let conn_string = format!("postgres://{postgres_user}:{postgres_password}@{postgres_address}?sslmode={postgres_sslmode}");
    let (postgres_client, connection): (tokio_postgres::Client, _) = tokio_postgres::connect(&conn_string, tokio_postgres::NoTls).await?;

    tokio::spawn(async move {
        if let Err(e) = connection.await {
            panic!("connection error: {}", e);
        }
    });

    let s3_bucket_name = match args.value_of("s3bucket") {
        Some(cli_bucket) => String::from(cli_bucket),
        None => match env::var("GEOCLOUD_S3_BUCKET") {
            Ok(env_bucket) => env_bucket,
            Err(_e) => String::from("geocloud")
        }
    };

    let s3_archive_bucket_name = match args.value_of("s3archivebucket") {
        Some(cli_bucket) => String::from(cli_bucket),
        None => match env::var("GEOCLOUD_S3_ARCHIVE_BUCKET") {
            Ok(env_bucket) => env_bucket,
            Err(_e) => String::from("")
        }
    };

    let s3_bucket: Bucket = Bucket::new_with_path_style(
        &s3_bucket_name, 
        Region::Custom {
            region: "".to_owned(),
            endpoint: "http://127.0.0.1:9000".to_owned()
        }, 
        Credentials::from_profile(Some("default"))?
    )?;

    let s3_archive_bucket: Bucket = Bucket::new(
        &s3_archive_bucket_name, 
        "us-east-1".parse()?, 
        Credentials::from_profile(Some("default"))?
    )?;

    let mut archive = Cursor::new(Vec::new());
    archive.write_all("job_id,task_type,job_status,job_error,start_time,end_time,job_args\n".as_bytes())?;
   
    let delete_before_minutes = match args.value_of_t("deletebefore") {
        Ok(delete_before) => delete_before,
        Err(_e) => 1440
    };


    let delete_before: DateTime<Utc> = Utc::now() - Duration::minutes(delete_before_minutes);
    println!("Cleaning up jobs before: {:?}", delete_before);
    for row in postgres_client.query("SELECT * FROM job", &[]).await? {
        let id: &str = row.try_get("job_id")?;
        let end_time: DateTime<Utc> = row.try_get("end_time")?;

        if delete_before > end_time {
            println!("Cleaning up Job ID: {}", id);
            let mut key = "jobs/".to_owned();
            key.push_str(id);
            key.push_str("/");
            let results = s3_bucket.list(key.clone(), None).await?;
            for result in results {
                for item in result.contents {
                    s3_bucket.delete_object(item.key).await?;
                }
            }

            s3_bucket.delete_object(key).await?;

            postgres_client.execute("DELETE FROM job WHERE job_id = $1", &[&id]).await?;

            dump_row_to_archive(row, &mut archive)?;
        }
    }

    archive.seek(SeekFrom::Start(0))?;
    let mut key = SystemTime::now().duration_since(SystemTime::UNIX_EPOCH)?.as_secs().to_string();
    key.push_str("/archive.csv");
    s3_archive_bucket.put_object_stream(& mut archive, key).await?;

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