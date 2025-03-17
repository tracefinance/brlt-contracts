package snowflake

import (
	"errors"
	"sync"
	"time"
)

// Constants defining the structure of the Snowflake ID
const (
	// epoch is set to January 1, 2020, 00:00:00 UTC in milliseconds
	epoch = 1577836800000

	// Bit lengths for each component
	sequenceBits     = 12
	machineIDBits    = 5
	dataCenterIDBits = 5
	timestampBits    = 41

	// Maximum values for each component
	maxDataCenterID = (1 << dataCenterIDBits) - 1 // 31
	maxMachineID    = (1 << machineIDBits) - 1    // 31
	sequenceMask    = (1 << sequenceBits) - 1     // 4095

	// Bit shifts for positioning each component in the 64-bit ID
	machineIDShift    = sequenceBits                         // 12
	dataCenterIDShift = machineIDShift + machineIDBits       // 17
	timestampShift    = dataCenterIDShift + dataCenterIDBits // 22
)

// Snowflake represents a generator for unique 64-bit IDs
type Snowflake struct {
	mutex         sync.Mutex // Ensures thread-safety for concurrent ID generation
	dataCenterID  int64      // Identifies the data center (0-31)
	machineID     int64      // Identifies the machine within the data center (0-31)
	sequence      int64      // Sequence number for IDs within the same millisecond
	lastTimestamp int64      // Last timestamp used to generate an ID
}

// NewSnowflake creates a new Snowflake instance with the given data center and machine IDs
func NewSnowflake(dataCenterID, machineID int64) (*Snowflake, error) {
	if dataCenterID < 0 || dataCenterID > maxDataCenterID {
		return nil, errors.New("dataCenterID must be between 0 and 31")
	}
	if machineID < 0 || machineID > maxMachineID {
		return nil, errors.New("machineID must be between 0 and 31")
	}
	return &Snowflake{
		dataCenterID:  dataCenterID,
		machineID:     machineID,
		sequence:      0,
		lastTimestamp: 0,
	}, nil
}

// currentTimestamp returns the current time in milliseconds since the Unix epoch
func currentTimestamp() int64 {
	return time.Now().UnixNano() / 1000000 // Convert nanoseconds to milliseconds
}

// GenerateID generates a unique 64-bit Snowflake ID
func (s *Snowflake) GenerateID() (int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	current := currentTimestamp()
	if current < epoch {
		return 0, errors.New("clock is set before the epoch")
	}
	if current < s.lastTimestamp {
		return 0, errors.New("clock moved backwards")
	}

	// Handle sequence within the same millisecond
	if current == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// Sequence overflow; wait for the next millisecond
			current = s.waitNextMillisecond()
			if current < s.lastTimestamp {
				return 0, errors.New("clock moved backwards while waiting for next millisecond")
			}
		}
	} else {
		// New millisecond; reset sequence
		s.sequence = 0
	}

	s.lastTimestamp = current

	// Construct the 64-bit ID using bit shifting and OR operations
	id := ((current - epoch) << timestampShift) |
		(s.dataCenterID << dataCenterIDShift) |
		(s.machineID << machineIDShift) |
		s.sequence

	return id, nil
}

// waitNextMillisecond waits until the next millisecond to avoid sequence overflow
func (s *Snowflake) waitNextMillisecond() int64 {
	current := currentTimestamp()
	for current <= s.lastTimestamp {
		current = currentTimestamp()
	}
	return current
}
