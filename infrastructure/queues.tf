variable "queues" {
    type        = list(string)
    default     = ["logsquaredn-geocloud"]
    description = "AWS SQS queue names to create"
}

variable "visibility_timeout_seconds" {
    type        = number
    default     = 15
    description = "Seconds for SQS to hide a message"

}

variable "message_retention_seconds" {
    type        = number
    default     = 1209600 # 14d (AWS specified max)
    description = "Seconds for SQS to retain a message"
}

resource "aws_sqs_queue" "queues" {
    for_each = toset(var.queues)

    name                       = each.key
    visibility_timeout_seconds = var.visibility_timeout_seconds
    message_retention_seconds  = var.message_retention_seconds
    tags                       = var.tags
}

output "queue_urls" {
    value = [for q in aws_sqs_queue.queues : q.url]
    description = "AWS SQS urls"
}

output "queue_names" {
    value = [for q in var.queues : q]
    description = "AWS SQS names"
}
