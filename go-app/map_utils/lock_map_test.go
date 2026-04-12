package map_utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLockMap(t *testing.T) {
	lm := NewLockMap[string, int]()
	assert.NotNil(t, lm, "Expected NewLockMap to return a non-nil pointer")
	assert.NotNil(t, lm.Map, "Expected Map to be initialized")
	assert.Lenf(t, lm.Map, 0, "Expected empty map, got length %d", len(lm.Map))
}

func TestLockMap_SetAndGet(t *testing.T) {
	lm := NewLockMap[string, int]()
	lm.Set("key1", 10)
	lm.Set("key2", 20)

	val, ok := lm.Get("key1")
	assert.True(t, ok, "Expected ok to be true")
	assert.Equalf(t, val, 10, "Expected key1=10, got val=%v", val)

	val, ok = lm.Get("key2")
	assert.True(t, ok, "Expected ok to be true")
	assert.Equalf(t, val, 20, "Expected key2=20, got val=%v", val)

	val, ok = lm.Get("nonexistent")
	assert.False(t, ok, "Expected nonexistent to return false")
}

func TestLockMap_Contains(t *testing.T) {
	lm := NewLockMap[string, int]()
	lm.Set("key1", 10)

	assert.True(t, lm.Contains("key1"), "Expected Contains('key1') to be true")
	assert.False(t, lm.Contains("key2"), "Expected Contains('key2') to be false")
}

func TestLockMap_Delete(t *testing.T) {
	lm := NewLockMap[string, int]()
	lm.Set("key1", 10)
	assert.True(t, lm.Contains("key1"), "Expected map to contain key1")

	lm.Delete("key1")
	assert.False(t, lm.Contains("key1"), "Expected map not to contain key1")
}

func TestLockMap_Clear(t *testing.T) {
	lm := NewLockMap[string, int]()
	lm.Set("key1", 10)
	lm.Set("key2", 20)
	assert.Lenf(t, lm.Map, 2, "Expected map to contain 2 items, got %v", len(lm.Map))

	lm.Clear()
	assert.Lenf(t, lm.Map, 0, "Expected map to contain no items, got %v", len(lm.Map))
}

func TestLockMap_ForEach(t *testing.T) {
	lm := NewLockMap[string, int]()
	lm.Set("a", 1)
	lm.Set("b", 2)
	lm.Set("c", 3)

	count := 0
	lm.ForEach(func(value int, key string) bool {
		count++
		return true
	})
	assert.Equal(t, 3, count, "Expected 3 iterations in ForEach, got %d", count)

	// Test early exit
	count = 0
	lm.ForEach(func(value int, key string) bool {
		count++
		return count < 2
	})
	assert.Equal(t, 2, count, "Expected ForEach to stop after 2 iterations, got %d", count)
}

func TestLockMap_ForEachMap(t *testing.T) {
	lm := NewLockMap[string, int]()
	lm.Set("a", 1)
	lm.Set("b", 2)

	lm.ForEachMap(func(value int, key string) int {
		return value * 10
	})

	valA, _ := lm.Get("a")
	assert.Equalf(t, 10, valA, "Expected a=10, got %d", valA)

	valB, _ := lm.Get("b")
	assert.Equalf(t, 20, valB, "Expected b=20, got %d", valB)
}

func TestLockMap_Mutate(t *testing.T) {
	t.Run("Replace", func(t *testing.T) {
		lm := NewLockMap[string, int]()
		lm.Set("a", 1)
		lm.Set("b", 2)

		lm.Mutate(func(value int, key string) LockMapMutateAction[string, int] {
			if key == "a" {
				return LockMapMutateAction[string, int]{
					Action: LockMapMutateActionType_Replace,
					Key:    "a",
					Value:  10,
				}
			}
			return LockMapMutateAction[string, int]{Action: LockMapMutateActionType_Noop}
		})

		valA, _ := lm.Get("a")
		assert.Equal(t, 10, valA, "Expected a=10 after Replace, got %d", valA)

		valB, _ := lm.Get("b")
		assert.Equal(t, 2, valB, "Expected b=2 to remain unchanged, got %d", valB)
	})

	t.Run("Delete", func(t *testing.T) {
		lm := NewLockMap[string, int]()
		lm.Set("a", 1)
		lm.Set("b", 2)

		lm.Mutate(func(value int, key string) LockMapMutateAction[string, int] {
			if key == "a" {
				return LockMapMutateAction[string, int]{
					Action: LockMapMutateActionType_Delete,
					Key:    "a",
				}
			}
			return LockMapMutateAction[string, int]{Action: LockMapMutateActionType_Noop}
		})

		assert.False(t, lm.Contains("a"), "Expected 'a' to be deleted via Mutate")
		assert.True(t, lm.Contains("b"), "Expected 'b' to still exist")
	})
}

func TestLockMap_Concurrency(t *testing.T) {
	lm := NewLockMap[int, int]()
	var wg sync.WaitGroup
	numGoroutines := 100
	iterations := 100

	wg.Add(numGoroutines * 2)

	// Writers
	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			for j := range iterations {
				lm.Set(id*iterations+j, j)
			}
		}(i)
	}

	// Readers
	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			for j := range iterations {
				lm.Get(id*iterations + j)
				lm.Contains(id*iterations + j)
			}
		}(i)
	}

	wg.Wait()
}
