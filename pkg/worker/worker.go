package worker

import (
	"context"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type Worker struct {
	ctx     context.Context
	service *service.Service
	queue   chan model.WorkerMessage
	logger  logger.Logger
}

func New(ctx context.Context, queue chan model.WorkerMessage, service *service.Service, logger logger.Logger) *Worker {
	return &Worker{
		ctx:     ctx,
		service: service,
		queue:   queue,
		logger:  logger,
	}
}

func (w *Worker) ProcessMessage() error {
	consumeErr := make(chan error, 1)
	go func() {
		for {
			if w.ctx.Err() != nil {
				consumeErr <- w.ctx.Err()
				return
			}
			message := <-w.queue
			switch message.Type {
			case bcModel.BasecampCommentMsg:
				_ = w.handleCommentMessage(w.logger, message.Payload)

			case bcModel.BasecampTodoMsg:
				_ = w.handleTodoMessage(w.logger, message.Payload)
			default:
				continue
			}
		}
	}()

	select {
	case err := <-consumeErr:
		return err
	case <-w.ctx.Done():
		return nil
	}
}

func (w *Worker) Enqueue(action string, msg interface{}) {
	w.queue <- model.WorkerMessage{Type: action, Payload: msg}
}

func (w *Worker) handleCommentMessage(l logger.Logger, payload interface{}) error {
	m := payload.(bcModel.BasecampCommentMessage)
	err := w.service.Basecamp.Comment.Create(m.ProjectID, m.RecordingID, m.Payload)
	if err != nil {
		l.Errorf(err, "failed to create basecamp comment", "payload", m.Payload.Content)
		return err
	}

	return nil
}

func (w *Worker) handleTodoMessage(l logger.Logger, payload interface{}) error {
	m := payload.(bcModel.BasecampTodoMessageModel)
	_, err := w.service.Basecamp.Todo.Create(m.ProjectID, m.ListID, m.Payload)
	if err != nil {
		l.Errorf(err, "failed to create basecamp todo", "payload", m.Payload.Content)
		return err
	}

	return nil
}
