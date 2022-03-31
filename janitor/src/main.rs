extern crate postgres;

use chrono::{DateTime, Utc};
use postgres::{Client, NoTls, Error};

fn main() -> Result<(), Error> {
    let mut client = Client::connect("postgres://geocloud:geocloud@localhost:5432?sslmode=disable", NoTls)?;

    for row in client.query("SELECT * FROM job", &[])? {
        let id: &str = row.try_get("job_id")?;
        let end_time: DateTime<Utc> = row.try_get("end_time")?;
        println!("{}", id);
        println!("{:?}", end_time);
    }

    Ok(())
}
