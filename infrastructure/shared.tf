terraform {
    backend "s3" {
        bucket = "logsquaredn-terraform-backends"
        key    = "geocloud/terraform.tfstate"
        region = "us-east-1"
    }
}

provider "aws" {
    region = var.region
}

variable "region" {
    type        = string
    default     = "us-east-1"
    description = "AWS region"
}

variable "tags" {
    type        = map(string)
    default     = {}
    description = "Tags to apply to resources"
}
