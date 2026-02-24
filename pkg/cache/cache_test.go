package cache

import (
	"testing"
	"time"
)

// TestInMemoryCacheSet tests setting and getting values
func TestInMemoryCacheSet(t *testing.T) {
	cache := NewInMemoryCache()

	cache.Set("key1", "value1", 1*time.Hour)
	value, exists := cache.Get("key1")

	if !exists {
		t.Error("Expected key to exist")
	}

	if value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}
}

// TestInMemoryCacheExpiration tests TTL expiration
func TestInMemoryCacheExpiration(t *testing.T) {
	cache := NewInMemoryCache()

	cache.Set("expire_key", "value", 100*time.Millisecond)

	// Immediate get should succeed
	value, exists := cache.Get("expire_key")
	if !exists {
		t.Error("Expected key to exist immediately after set")
	}

	if value != "value" {
		t.Errorf("Expected 'value', got %v", value)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Get after expiration should fail
	_, exists = cache.Get("expire_key")
	if exists {
		t.Error("Expected key to be expired")
	}
}

// TestInMemoryCacheDelete tests deleting values
func TestInMemoryCacheDelete(t *testing.T) {
	cache := NewInMemoryCache()

	cache.Set("del_key", "value", 1*time.Hour)
	cache.Delete("del_key")

	_, exists := cache.Get("del_key")
	if exists {
		t.Error("Expected key to be deleted")
	}
}

// TestInMemoryCacheClear tests clearing all values
func TestInMemoryCacheClear(t *testing.T) {
	cache := NewInMemoryCache()

	cache.Set("key1", "value1", 1*time.Hour)
	cache.Set("key2", "value2", 1*time.Hour)
	cache.Set("key3", "value3", 1*time.Hour)

	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}
}

// TestInMemoryCacheSize tests size tracking
func TestInMemoryCacheSize(t *testing.T) {
	cache := NewInMemoryCache()

	if cache.Size() != 0 {
		t.Errorf("Expected initial size 0, got %d", cache.Size())
	}

	cache.Set("key1", "value1", 1*time.Hour)
	if cache.Size() != 1 {
		t.Errorf("Expected size 1 after first set, got %d", cache.Size())
	}

	cache.Set("key2", "value2", 1*time.Hour)
	if cache.Size() != 2 {
		t.Errorf("Expected size 2 after second set, got %d", cache.Size())
	}

	cache.Delete("key1")
	if cache.Size() != 1 {
		t.Errorf("Expected size 1 after delete, got %d", cache.Size())
	}
}

// TestInMemoryCacheExists tests existence checking
func TestInMemoryCacheExists(t *testing.T) {
	cache := NewInMemoryCache()

	cache.Set("exists_key", "value", 1*time.Hour)

	if !cache.Exists("exists_key") {
		t.Error("Expected exists_key to exist")
	}

	if cache.Exists("nonexistent") {
		t.Error("Expected nonexistent key to not exist")
	}

	cache.Delete("exists_key")
	if cache.Exists("exists_key") {
		t.Error("Expected exists_key to not exist after delete")
	}
}

// TestInMemoryCacheStats tests statistics tracking
func TestInMemoryCacheStats(t *testing.T) {
	cache := NewInMemoryCache()

	cache.Set("key1", "value1", 1*time.Hour)
	cache.Get("key1") // Hit
	cache.Get("key1") // Hit
	cache.Get("nonexistent") // Miss

	stats := cache.GetStats()

	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}

	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}

	if stats.Total != 1 {
		t.Errorf("Expected 1 total item, got %d", stats.Total)
	}
}

// TestInMemoryCacheGetEntry tests retrieving full entry
func TestInMemoryCacheGetEntry(t *testing.T) {
	cache := NewInMemoryCache()

	cache.Set("entry_key", "entry_value", 1*time.Hour)

	entry, exists := cache.GetEntry("entry_key")
	if !exists {
		t.Error("Expected entry to exist")
	}

	if entry == nil {
		t.Error("Expected entry to not be nil")
		return
	}

	if entry.Value != "entry_value" {
		t.Errorf("Expected 'entry_value', got %v", entry.Value)
	}

	if entry.HitCount != 0 {
		t.Errorf("Expected HitCount 0, got %d", entry.HitCount)
	}

	// Access the value to increment hit count
	cache.Get("entry_key")

	entry, _ = cache.GetEntry("entry_key")
	if entry.HitCount != 1 {
		t.Errorf("Expected HitCount 1 after access, got %d", entry.HitCount)
	}
}

// TestInMemoryCacheMultipleTypes tests storing different types
func TestInMemoryCacheMultipleTypes(t *testing.T) {
	cache := NewInMemoryCache()

	// String
	cache.Set("string_key", "string_value", 1*time.Hour)

	// Integer
	cache.Set("int_key", 42, 1*time.Hour)

	// Slice
	cache.Set("slice_key", []string{"a", "b", "c"}, 1*time.Hour)

	// Map
	testMap := map[string]interface{}{"nested": "value"}
	cache.Set("map_key", testMap, 1*time.Hour)

	// Verify string
	val, _ := cache.Get("string_key")
	if val != "string_value" {
		t.Errorf("Expected 'string_value', got %v", val)
	}

	// Verify int
	val, _ = cache.Get("int_key")
	if val != 42 {
		t.Errorf("Expected 42, got %v", val)
	}

	// Verify slice
	val, _ = cache.Get("slice_key")
	slice := val.([]string)
	if len(slice) != 3 || slice[0] != "a" {
		t.Errorf("Expected 3-element slice, got %v", val)
	}

	// Verify map
	val, _ = cache.Get("map_key")
	m := val.(map[string]interface{})
	if m["nested"] != "value" {
		t.Errorf("Expected nested value, got %v", m)
	}
}

// TestCacheWithTTL tests the TTL wrapper
func TestCacheWithTTL(t *testing.T) {
	base := NewInMemoryCache()
	cache := NewCacheWithTTL(base, 1*time.Hour)

	// Set with default TTL
	cache.Set("key1", "value1")
	value, exists := cache.Get("key1")

	if !exists {
		t.Error("Expected key to exist")
	}

	if value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}

	// Set with custom TTL
	cache.SetWithTTL("key2", "value2", 100*time.Millisecond)
	value, exists = cache.Get("key2")

	if !exists {
		t.Error("Expected key2 to exist")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	_, exists = cache.Get("key2")
	if exists {
		t.Error("Expected key2 to be expired")
	}

	// key1 should still exist (1 hour TTL)
	_, exists = cache.Get("key1")
	if !exists {
		t.Error("Expected key1 to still exist")
	}
}

// TestInMemoryCacheConcurrency tests concurrent access
func TestInMemoryCacheConcurrency(t *testing.T) {
	cache := NewInMemoryCache()

	// Set values concurrently
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			key := "key_" + string(rune(48+index))
			cache.Set(key, index, 1*time.Hour)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if cache.Size() != 10 {
		t.Errorf("Expected size 10, got %d", cache.Size())
	}

	// Read values concurrently
	for i := 0; i < 10; i++ {
		go func(index int) {
			key := "key_" + string(rune(48+index))
			cache.Get(key)
			done <- true
		}(i)
	}

	// Wait for all reads
	for i := 0; i < 10; i++ {
		<-done
	}

	// Stats should show many reads
	stats := cache.GetStats()
	if stats.Hits < 10 {
		t.Errorf("Expected at least 10 hits, got %d", stats.Hits)
	}
}

// BenchmarkCacheGet benchmarks cache Get operation
func BenchmarkCacheGet(b *testing.B) {
	cache := NewInMemoryCache()
	cache.Set("bench_key", "bench_value", 1*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("bench_key")
	}
}

// BenchmarkCacheSet benchmarks cache Set operation
func BenchmarkCacheSet(b *testing.B) {
	cache := NewInMemoryCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("bench_key", "bench_value", 1*time.Hour)
	}
}
