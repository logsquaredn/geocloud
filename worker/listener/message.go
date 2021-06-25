package listener

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/logsquaredn/geocloud/worker"
)

type SQSMessage struct {
	msg *sqs.Message 
}

var _ worker.Message = (*SQSMessage)(nil)
