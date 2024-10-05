package cache_test

import (
	"context"
	cache2 "github.com/debugger84/sqlc-dataloader/cache"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestLRU_Set(t *testing.T) {

	type User struct {
		ID        int
		Email     string
		FirstName string
		LastName  string
	}

	m := map[int]*User{
		5: &User{ID: 5, FirstName: "John", LastName: "Smith", Email: "john@example.com"},
	}

	batchFunc := func(_ context.Context, keys []int) []*dataloader.Result[*User] {
		var results []*dataloader.Result[*User]
		// do some pretend work to resolve keys
		for _, k := range keys {
			results = append(results, &dataloader.Result[*User]{Data: m[k]})
		}
		return results
	}

	cache := cache2.NewLRU[int, *User](100, time.Minute)
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithCache[int, *User](cache))

	// immediately call the future function from loader
	result, err := loader.Load(context.TODO(), 5)()

	assert.NoError(t, err)
	assert.Equal(t, m[5], result)
	cachedItem, success := cache.Get(context.Background(), 5)
	assert.True(t, success)
	val, err := cachedItem()
	require.NotNil(t, val)
	assert.NoError(t, err)
	assert.Equal(t, 5, val.ID)
}

func TestLRU_GetFromCache(t *testing.T) {

	type User struct {
		ID        int
		Email     string
		FirstName string
		LastName  string
	}

	m := map[int]*User{
		5: &User{ID: 5, FirstName: "John", LastName: "Smith", Email: "john@example.com"},
	}

	counter := 0
	batchFunc := func(_ context.Context, keys []int) []*dataloader.Result[*User] {
		var results []*dataloader.Result[*User]
		counter++
		// do some pretend work to resolve keys
		for _, k := range keys {
			results = append(results, &dataloader.Result[*User]{Data: m[k]})
		}
		return results
	}

	cache := cache2.NewLRU[int, *User](100, time.Minute)
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithCache[int, *User](cache))

	// immediately call the future function from loader
	result1, err1 := loader.Load(context.TODO(), 5)()
	result2, err2 := loader.Load(context.TODO(), 5)()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, m[5], result1)
	assert.Equal(t, m[5], result2)
	assert.Equal(t, 1, counter)
}

func TestLRU_GetBatch(t *testing.T) {
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

	cache := cache2.NewLRU[int, *User](100, time.Minute)
	loader := dataloader.NewBatchedLoader(
		batchFunc,
		dataloader.WithCache[int, *User](cache),
		dataloader.WithBatchCapacity[int, *User](10),
		dataloader.WithWait[int, *User](16*time.Millisecond),
	)

	var result1, result2 *User
	var err1, err2 error
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		result1, err1 = loader.Load(context.TODO(), 5)()
	}()

	go func() {
		defer wg.Done()
		result2, err2 = loader.Load(context.TODO(), 10)()
	}()

	wg.Wait()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, m[5], result1)
	assert.Equal(t, m[10], result2)
	assert.Equal(t, 1, counter)
	assert.Equal(t, 2, keysCount)
}
