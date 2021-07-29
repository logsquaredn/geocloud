variable "bucket" {
    type        = string
    default     = "logsquaredn-geocloud"
    description = "AWS S3 bucket to create"
}

resource "aws_s3_bucket" "bucket" {
    bucket = var.bucket
    acl    = "private"
    tags   = var.tags
}

output "bucket" {
    value = aws_s3_bucket.bucket.bucket
    description = "AWS S3 bucket"
}
