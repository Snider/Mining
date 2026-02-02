package node

// MessageDispatcher routes messages to either Controller or Worker
type MessageDispatcher struct {
	controller *Controller
	worker     *Worker
}

// NewMessageDispatcher creates a new MessageDispatcher
func NewMessageDispatcher(controller *Controller, worker *Worker) *MessageDispatcher {
	return &MessageDispatcher{
		controller: controller,
		worker:     worker,
	}
}

// Dispatch routes incoming messages to the appropriate handler
func (d *MessageDispatcher) Dispatch(conn *PeerConnection, msg *Message) {
	// If ReplyTo is set, it's a response to a request initiated by Controller
	if msg.ReplyTo != "" && d.controller != nil {
		d.controller.HandleResponse(conn, msg)
		return
	}

	// Otherwise treat it as a new request for Worker
	if d.worker != nil {
		d.worker.HandleMessage(conn, msg)
	}
}
