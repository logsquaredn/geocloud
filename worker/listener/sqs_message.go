package listener

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/logsquaredn/geocloud"
)

type SQSMessage struct {
	msg *sqs.Message 
}

var _ geocloud.Message = (*SQSMessage)(nil)

func (m *SQSMessage) ID() string {
	return *m.msg.Body
}
