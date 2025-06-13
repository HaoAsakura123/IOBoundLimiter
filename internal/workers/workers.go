package workers

import (
	"context"
	"fmt"
	"ioboundlimiter/internal/storage"
	"log"
	"math/rand"
	"sync"
	"time"
)

var (
	tasksChan   chan string
	semaphore   chan struct{}
	wg          sync.WaitGroup // Для ожидания завершения воркеров
	shutdownCtx context.Context
	cancelFunc  context.CancelFunc
)

const maxWorkers = 5

func InitWorkers() {

	shutdownCtx, cancelFunc = context.WithCancel(context.Background())
	tasksChan = make(chan string, 100)
	semaphore = make(chan struct{}, maxWorkers)

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(i)
	}
}

func Shutdown() {
	cancelFunc()
	close(tasksChan)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All workers stopped gracefully")
	case <-time.After(5 * time.Second):
		log.Println("Timeout: some workers did not finish")
	}
}

func worker(id int) {
	defer wg.Done()

	for {
		select {
		case <-shutdownCtx.Done():
			log.Printf("Worker %d: shutting down...", id)
			return
		case uuid, ok := <-tasksChan:
			if !ok {
				log.Printf("Worker %d: no more tasks, exiting", id)
				return
			}

			semaphore <- struct{}{}
			if !storage.IsExists(uuid) {
				<-semaphore
				continue
			}

			if err := processTask(id, uuid); err != nil {
				log.Printf("Worker %d: task %s failed: %v", id, uuid, err)
			}
			<-semaphore
		}
	}
}

func processTask(id int, uuid string) error {
	status := fmt.Sprintf("Worker %d starting task: %s", id, uuid)
	if err := usefulWork(uuid, status, id); err != nil {
		return err
	}
	time.Sleep(time.Duration(rand.Intn(40)+60) * time.Second)

	status = fmt.Sprintf("Worker %d asks BD while working with: %s", id, uuid)
	if err := usefulWork(uuid, status, id); err != nil {
		return err
	}
	time.Sleep(time.Duration(rand.Intn(40)+60) * time.Second)

	status = fmt.Sprintf("Worker %d sends other bd results about working task: %s", id, uuid)
	if err := usefulWork(uuid, status, id); err != nil {
		return err
	}

	log.Printf("Worker %d success ends task: %s", id, uuid)
	return nil
}

func AddToChannel(uuid string) error {
	select {
	case tasksChan <- uuid:
		log.Printf("task received: %s", uuid)
	default:
		return fmt.Errorf("cannot add task: %s (channel full)", uuid)
	}
	return nil
}

func usefulWork(task, status string, id int) error {
	if err := storage.ChangeStatus(task, status); err != nil {
		log.Printf("Worker %d ends task: %s", id, task)
		return err
	}
	log.Print(status)
	return nil
}
