package http_logger

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-slark/slark/logger"
	"net/http"
	"time"
)

type LoggerForwardingQueue struct {
	Intake        chan Log
	logger        logger.Logger
	retryInterval time.Duration
}

func NewLoggerForwardingQueue(conf AccessLoggerConfig) (q *LoggerForwardingQueue) {
	return &LoggerForwardingQueue{
		Intake: make(chan Log, conf.DropSize),
		logger: conf.Logger,
	}
}

func (q *LoggerForwardingQueue) intake() chan Log {
	return q.Intake
}

func (q *LoggerForwardingQueue) run() {
	// Forwards payloads asynchronously
	for {
		logEntry := <-q.Intake
		payload := buildPayload(&logEntry)

		ctx := context.TODO()

		// Let's convert our fields to their JSON counterparts before logging fields as JSON
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			q.logger.Log(ctx, logger.ErrorLevel, map[string]interface{}{}, fmt.Sprintf("Impossible to marshal log payload to JSON: %v (payload: %v)", err, payload))
			continue
		}
		var payloadJSON map[string]interface{}
		err = json.Unmarshal(payloadBytes, &payloadJSON)
		if err != nil {
			q.logger.Log(ctx, logger.ErrorLevel, map[string]interface{}{}, fmt.Sprintf("Impossible to unmarshal log payload into map[string]interface{}: %v (payload: %s)", err, payloadBytes))
			continue
		}

		// Let's forward the log line to fluentd
		//logger := q.logrusLogger.WithFields(payloadJSON)
		if payload.Response.Status >= http.StatusInternalServerError {
			q.logger.Log(ctx, logger.ErrorLevel, payloadJSON, "server error")
		} else if payload.Response.Status >= http.StatusBadRequest {
			q.logger.Log(ctx, logger.WarnLevel, payloadJSON, "client error")
		} else {
			q.logger.Log(ctx, logger.InfoLevel, payloadJSON, "request processed")
		}
	}
}
