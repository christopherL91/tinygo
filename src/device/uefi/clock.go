//go:build uefi

package uefi

import "sync"

var calibrateMutex sync.Mutex
var calculatedFrequency uint64

// TicksFrequency returns the frequency of Ticks, in Hz.
//
// It must not defer: the scheduler calls it, through the runtime's
// nanosecondsToTicks, to see whether a timer is due, and it does so outside
// any goroutine, where a function with a defer has no task to keep its defer
// frame in.
func TicksFrequency() uint64 {
	frequency := getTSCFrequency()
	if frequency > 0 {
		return frequency
	}

	calibrateMutex.Lock()
	if calculatedFrequency == 0 {
		calculatedFrequency = calibrateFrequency()
	}
	frequency = calculatedFrequency
	calibrateMutex.Unlock()
	return frequency
}

// calibrateFrequency counts ticks across a quarter of a second of the
// firmware's timer.
func calibrateFrequency() uint64 {
	var event EFI_EVENT
	var index UINTN
	if BS().CreateEvent(EVT_TIMER, TPL_CALLBACK, nil, nil, &event) != EFI_SUCCESS {
		return 0
	}

	var frequency uint64
	start := Ticks()
	if BS().SetTimer(event, TimerPeriodic, 250*10000) == EFI_SUCCESS &&
		BS().WaitForEvent(1, &event, &index) == EFI_SUCCESS {
		frequency = (Ticks() - start) * 4
	}
	BS().CloseEvent(event)
	return frequency
}
