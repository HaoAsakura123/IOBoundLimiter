package storage

import (
	"fmt"
	"ioboundlimiter/internal/util"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ioBound     = make(map[string]Status)
	lockIOBound = &sync.RWMutex{}
)

type Status struct {
	CurStatus  string    `json:"status"`
	DateCreate time.Time `json:"date"`
	Name       string    `json:"name"`

	DateOutput string `json:"dateout"`
}

func AddToStorage(nameTask string) (string, error) {
	if nameTask == ""{
		log.Printf("task name is empty")
		return "", fmt.Errorf("task name is empty")
	}
	if IsExists(nameTask) {
		log.Printf("task %s already exists", nameTask)
		return "", fmt.Errorf("task %s already exists", nameTask)
	}

	currTime := util.TimeNow()

	dateOut, err := util.TimeFormat()
	if err != nil {
		log.Printf("something go wrong: %v", err)
		return "", fmt.Errorf("something go wrong: %v", err)
	}

	uuid := uuid.New().String()

	stat := Status{CurStatus: "pending", DateCreate: currTime, Name: nameTask, DateOutput: dateOut}

	if err := setTask(stat, uuid); err != nil {
		log.Printf("Something go wrong with setting task: %v", err)
		return "", fmt.Errorf("something go wrong: %v", err)
	}

	return uuid, nil
}

func IsExists(uuid string) bool {
	lockIOBound.RLock()
	_, exists := ioBound[uuid]
	lockIOBound.RUnlock()

	return exists
}

func setTask(stat Status, uuid string) error {
	if IsExists(uuid) {
		log.Printf("cannot create task with this UUID: %s", uuid)
		return fmt.Errorf("cannot create task with this UUID: %s", uuid)
	}

	lockIOBound.Lock()
	ioBound[uuid] = stat
	lockIOBound.Unlock()

	return nil
}

func ChangeStatus(uuid, status string) error {
	if !IsExists(uuid) {
		return fmt.Errorf("task %s is not exists", uuid)
	}

	lockIOBound.Lock()

	stat := ioBound[uuid]
	stat.CurStatus = status
	ioBound[uuid] = stat

	lockIOBound.Unlock()

	return nil
}

func DeleteTask(uuid string) error {
	if !IsExists(uuid) {
		return fmt.Errorf("task %s doesnt exist", uuid)
	}

	delete(ioBound, uuid)

	return nil
}

func GetResponse(uuid string) (Status, error) {
	if !IsExists(uuid) {
		return Status{}, fmt.Errorf("task %s doesnt exists", uuid)
	}
	lockIOBound.RLock()
	response := ioBound[uuid]
	lockIOBound.RUnlock()

	return response, nil
}
