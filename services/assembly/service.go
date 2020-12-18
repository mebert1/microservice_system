package assembly

import (
	"encoding/json"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/rbmq"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/service"
	"go.uber.org/zap"
)

// Service is the instance wrapper
type Service struct {
	*service.Service
}

// New is the initializer
func New(config *service.Config, messages chan rbmq.Message, logger *zap.SugaredLogger) (*Service, error) {
	var err error

	supplierService := &Service{}
	supplierService.Service, err = service.New(config, messages, logger)
	if err != nil {
		return nil, err
	}

	go supplierService.handleRbmqMessage(messages)

	return supplierService, nil
}

/* handleRbmqMessage receives incoming messages from factory service */
/* after reception, function sleeps to simulate production and then responds to factory with ack msg */
func (s *Service) handleRbmqMessage(messages <-chan rbmq.Message) {
	for msg := range messages {

		recMsg := rbmq.OrderMessage{}
		err := json.Unmarshal(msg.Body, &recMsg)
		if err != nil {
			s.Logger.Errorw("Failed to parse message", "err", err, "msg", string(msg.Body))
			continue
		}

		s.Logger.Infow("Received assembly request", "order", recMsg.OrderID)
		s.Logger.Infow("Starting production", "order", recMsg.OrderID)

		/* sleep to simulate production process */
		/* sleep duration depends on service location and individual product */
		for _, item := range recMsg.Items {
			produce(item.AssemblyTime, s.Config.Location)
			s.Logger.Infow("Successfully produced item", "order", recMsg.OrderID, "item", item.ItemID, "assemblyTime", item.AssemblyTime)
		}
		s.Logger.Infow("Production finished", "order", recMsg.OrderID)

		/* Notify factory service of finished assembling process */
		response, err := productionAck(recMsg)
		if err != nil {
			s.Logger.Errorw("Failed to marshal assembly acknowledgement", "err", err)
			return
		}
		s.Producer[s.Config.Location].Publish(response, "factory")
	}
}

// TODO: Add equation to depend sleep duration on location and individual product
func produce(assemblyTime int, location string) {
	var productionTime float32
	if location == "usa" {
		productionTime = float32(assemblyTime) * float32(0.7)
	} else if location == "china" {
		productionTime = float32(assemblyTime) * float32(1.2)
	}

	time.Sleep(time.Duration(productionTime) * time.Second)
}

func productionAck(recMsg rbmq.OrderMessage) ([]byte, error) {

	response := rbmq.OrderMessage{
		Timestamp: time.Now().UTC(),
		MsgType:   recMsg.MsgType,
		Status:    "complete",
		OrderID:   recMsg.OrderID,
	}

	responseBody, err := json.Marshal(response)

	return responseBody, err
}
