package db

import (
	"sync"
	"testing"
	"time"
)

// Helper functions to extract components from a Snowflake ID
func extractTimestamp(id int64) int64 {
	return (id >> timestampShift) + epoch
}

func extractDataCenterID(id int64) int64 {
	return (id >> dataCenterIDShift) & 0x1F // 5 bits
}

func extractMachineID(id int64) int64 {
	return (id >> machineIDShift) & 0x1F // 5 bits
}

func extractSequence(id int64) int64 {
	return id & sequenceMask // 12 bits
}

// TestBasicFunctionality ensures IDs are unique and increasing
func TestBasicFunctionality(t *testing.T) {
	sf, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("Failed to create Snowflake: %v", err)
	}
	ids := make(map[int64]bool)
	var prevID int64
	for i := 0; i < 100; i++ {
		id, err := sf.GenerateID()
		if err != nil {
			t.Errorf("Failed to generate ID: %v", err)
		}
		if ids[id] {
			t.Errorf("Duplicate ID generated: %d", id)
		}
		ids[id] = true
		if i > 0 && id <= prevID {
			t.Errorf("IDs are not increasing: %d <= %d", id, prevID)
		}
		prevID = id
	}
}

// TestInvalidInputs checks error handling for invalid inputs
func TestInvalidInputs(t *testing.T) {
	// Test negative dataCenterID
	_, err := NewSnowflake(-1, 1)
	if err == nil {
		t.Error("Expected error for negative dataCenterID")
	}

	// Test negative machineID
	_, err = NewSnowflake(1, -1)
	if err == nil {
		t.Error("Expected error for negative machineID")
	}

	// Test out-of-range dataCenterID (max 31 for 5 bits)
	_, err = NewSnowflake(32, 1)
	if err == nil {
		t.Error("Expected error for out-of-range dataCenterID")
	}

	// Test out-of-range machineID (max 31 for 5 bits)
	_, err = NewSnowflake(1, 32)
	if err == nil {
		t.Error("Expected error for out-of-range machineID")
	}
}

// TestDataCenterAndMachineID verifies embedded IDs
func TestDataCenterAndMachineID(t *testing.T) {
	sf, err := NewSnowflake(1, 2)
	if err != nil {
		t.Fatalf("Failed to create Snowflake: %v", err)
	}
	id, err := sf.GenerateID()
	if err != nil {
		t.Errorf("Failed to generate ID: %v", err)
	}
	dcID := extractDataCenterID(id)
	mID := extractMachineID(id)
	if dcID != 1 {
		t.Errorf("Expected dataCenterID 1, got %d", dcID)
	}
	if mID != 2 {
		t.Errorf("Expected machineID 2, got %d", mID)
	}
}

// TestSequenceHandling checks sequence increment within the same millisecond
func TestSequenceHandling(t *testing.T) {
	sf, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("Failed to create Snowflake: %v", err)
	}
	var prevID int64
	for i := 0; i < 10; i++ {
		id, err := sf.GenerateID()
		if err != nil {
			t.Errorf("Failed to generate ID: %v", err)
		}
		if i > 0 {
			prevSeq := extractSequence(prevID)
			currSeq := extractSequence(id)
			prevTs := extractTimestamp(prevID)
			currTs := extractTimestamp(id)
			if currTs == prevTs {
				if currSeq != prevSeq+1 {
					t.Errorf("Sequence did not increment correctly: %d -> %d", prevSeq, currSeq)
				}
			} else {
				if currSeq != 0 {
					t.Errorf("Sequence did not reset on timestamp change: %d", currSeq)
				}
			}
		}
		prevID = id
	}
}

// TestTimestampHandling ensures timestamps increase and sequences reset
func TestTimestampHandling(t *testing.T) {
	sf, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("Failed to create Snowflake: %v", err)
	}
	id1, err := sf.GenerateID()
	if err != nil {
		t.Errorf("Failed to generate ID: %v", err)
	}
	time.Sleep(2 * time.Millisecond) // Ensure timestamp changes
	id2, err := sf.GenerateID()
	if err != nil {
		t.Errorf("Failed to generate ID: %v", err)
	}
	ts1 := extractTimestamp(id1)
	ts2 := extractTimestamp(id2)
	if ts2 <= ts1 {
		t.Errorf("Timestamp did not increase: %d <= %d", ts2, ts1)
	}
	seq2 := extractSequence(id2)
	if seq2 != 0 {
		t.Errorf("Sequence did not reset on timestamp change: %d", seq2)
	}
}

// TestConcurrency verifies thread-safety and uniqueness under concurrent generation
func TestConcurrency(t *testing.T) {
	sf, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("Failed to create Snowflake: %v", err)
	}
	var wg sync.WaitGroup
	ids := make(chan int64, 1000)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				id, err := sf.GenerateID()
				if err != nil {
					t.Errorf("Error generating ID: %v", err)
				}
				ids <- id
			}
		}()
	}
	wg.Wait()
	close(ids)
	idSet := make(map[int64]bool)
	for id := range ids {
		if idSet[id] {
			t.Errorf("Duplicate ID generated: %d", id)
		}
		idSet[id] = true
	}
}
