package storage

import (
	"ioboundlimiter/internal/util"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddToStorage(t *testing.T) {
    t.Run("tasks with same name but different UUIDs", func(t *testing.T) {
        taskName := "same_name_task"
        
        id1, err := AddToStorage(taskName)
        assert.NoError(t, err)
        assert.NotEmpty(t, id1)
        
        id2, err := AddToStorage(taskName)
        assert.NoError(t, err) 
        assert.NotEmpty(t, id2)
        
        assert.NotEqual(t, id1, id2)
        
        assert.True(t, IsExists(id1))
        assert.True(t, IsExists(id2))
    })
}

func TestIsExists(t *testing.T) {
	t.Run("existing task", func(t *testing.T) {
		id, _ := AddToStorage("existing_task")
		assert.True(t, IsExists(id))
	})

	t.Run("non-existing task", func(t *testing.T) {
		assert.False(t, IsExists(uuid.New().String()))
	})
}

func TestSetTask(t *testing.T) {
	t.Run("set new task", func(t *testing.T) {
		id := uuid.New().String()
		status := Status{
			CurStatus:  "processing",
			Name:       "new_task",
			DateCreate: util.TimeNow(),
			DateOutput: time.Now().Format(time.RFC3339),
		}

		err := setTask(status, id)
		assert.NoError(t, err)
		assert.True(t, IsExists(id))
	})

	t.Run("set duplicate task", func(t *testing.T) {
		id, _ := AddToStorage("duplicate_test")
		status := Status{
			CurStatus:  "processing",
			Name:       "duplicate_test",
			DateCreate: util.TimeNow(),
		}

		err := setTask(status, id)
		assert.Error(t, err)
	})
}

func TestChangeStatus(t *testing.T) {
	t.Run("valid status change", func(t *testing.T) {
		id, _ := AddToStorage("status_change_test")
		err := ChangeStatus(id, "completed")
		assert.NoError(t, err)

		task, _ := GetResponse(id)
		assert.Equal(t, "completed", task.CurStatus)
	})

	t.Run("non-existent task", func(t *testing.T) {
		err := ChangeStatus(uuid.New().String(), "completed")
		assert.Error(t, err)
	})
}

func TestDeleteTask(t *testing.T) {
	t.Run("delete existing task", func(t *testing.T) {
		id, _ := AddToStorage("to_delete")
		assert.True(t, IsExists(id))

		err := DeleteTask(id)
		assert.NoError(t, err)
		assert.False(t, IsExists(id))
	})

	t.Run("delete non-existent task", func(t *testing.T) {
		err := DeleteTask(uuid.New().String())
		assert.Error(t, err)
	})
}

func TestGetResponse(t *testing.T) {
	t.Run("get existing task", func(t *testing.T) {
		taskName := "get_test_task"
		id, _ := AddToStorage(taskName)

		task, err := GetResponse(id)
		assert.NoError(t, err)
		assert.Equal(t, taskName, task.Name)
		assert.Equal(t, "pending", task.CurStatus)
	})

	t.Run("get non-existent task", func(t *testing.T) {
		_, err := GetResponse(uuid.New().String())
		assert.Error(t, err)
	})
}

func TestConcurrentAccess(t *testing.T) {
	const numWorkers = 100
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	t.Run("concurrent task addition", func(t *testing.T) {
		for i := 0; i < numWorkers; i++ {
			go func(i int) {
				defer wg.Done()
				_, err := AddToStorage("concurrent_task")
				assert.NoError(t, err)
			}(i)
		}
		wg.Wait()
	})

	t.Run("concurrent status changes", func(t *testing.T) {
		id, _ := AddToStorage("concurrent_status_test")
		wg.Add(numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func(i int) {
				defer wg.Done()
				err := ChangeStatus(id, "processing")
				assert.NoError(t, err)
			}(i)
		}
		wg.Wait()

		task, _ := GetResponse(id)
		assert.Equal(t, "processing", task.CurStatus)
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty task name", func(t *testing.T) {
		_, err := AddToStorage("")
		assert.Error(t, err)
	})

	t.Run("long task name", func(t *testing.T) {
		longName := make([]byte, 1000)
		_, err := AddToStorage(string(longName))
		assert.NoError(t, err)
	})
}