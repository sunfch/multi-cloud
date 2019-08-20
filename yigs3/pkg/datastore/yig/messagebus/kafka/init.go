package kafka

import (
	bus "github.com/opensds/multi-cloud/yigs3/pkg/datastore/yig/messagebus"
	"github.com/opensds/multi-cloud/yigs3/pkg/datastore/yig/messagebus/types"
)

func init() {
	kafkaBuilder := &KafkaBuilder{}

	bus.AddMsgBuilder(types.MSG_BUS_SENDER_KAFKA, kafkaBuilder)
}
