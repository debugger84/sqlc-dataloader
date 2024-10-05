package cache_test

import (
	"context"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestNoCache_GetBatch(t *testing.T) {
	type User struct {
		ID        int
		Email     string
		FirstName string
		LastName  string
	}

	m := map[int]*User{
		5:  &User{ID: 5, FirstName: "John", LastName: "Smith", Email: "john@example.com"},
		10: &User{ID: 5, FirstName: "Jeniffer", LastName: "Smith", Email: "jeniffer@example.com"},
	}

	keysCount := 0
	counter := 0
	batchFunc := func(_ context.Context, keys []int) []*dataloader.Result[*User] {
		var results []*dataloader.Result[*User]

		counter++
		keysCount = len(keys)
		// do some pretend work to resolve keys
		for _, k := range keys {
			results = append(results, &dataloader.Result[*User]{Data: m[k]})
		}
		return results
	}

	cache := &dataloader.NoCache[int, *User]{}
	loader := dataloader.NewBatchedLoader(
		batchFunc,
		dataloader.WithCache[int, *User](cache),
		dataloader.WithBatchCapacity[int, *User](10),
		dataloader.WithWait[int, *User](16*time.Millisecond),
	)

	var result1, result2 *User
	var err1, err2 error
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		result1, err1 = loader.Load(context.TODO(), 5)()
	}()

	go func() {
		defer wg.Done()
		result2, err2 = loader.Load(context.TODO(), 10)()
	}()
	go func() {
		defer wg.Done()
		result1, err1 = loader.Load(context.TODO(), 5)()
	}()

	wg.Wait()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, m[5], result1)
	assert.Equal(t, m[10], result2)
	assert.Equal(t, 1, counter)
	assert.Equal(t, 3, keysCount)
}
