package task

import "github.com/chistyakoviv/converter/internal/service"

type serv struct {
	queue chan interface{}
}

func NewService() service.TaskService {
	return &serv{
		// The size of channel is 2 to check for new tasks immediately
		// after some task is processed
		queue: make(chan interface{}, 2),
	}
}

// Try to add a task only if the queue is not full
func (s *serv) TrySchedule() bool {
	select {
	case s.queue <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *serv) Tasks() <-chan interface{} {
	return s.queue
}
