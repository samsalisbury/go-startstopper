// Package startstopper defines a single type StartStopper which can be used to
// signal that a process has completed (similar to closing a channel). Then at
// a later time, the process can be restarted, by effectively "reopening" the
// channel.
//
// Of course, reopening a channel is not possible in Go, for good reason.
// This is because channels are often concerned with transmitting data, so we
// need to be sure that once a channel is closed, there will be no more data
// coming.
//
// The other major use of channels is in signalling completion, of a task, or
// otherwise that something should stop. In this instance, it is often clean and
// natural to simply end any interested goroutines, and have new ones created
// later if the process needs to be restarted. In that case, you do not need
// this library, since existing close(channel) semantics work fine for that use
// case.
//
// Where this library becomes useful is if you have a long-lived service
// instance that must go into a dormant state for some time, and later become
// active again.
package startstopper

import "sync"

// StartStopper can be used in place of close(chan) to signal that something has
// finished or stopped. It adds the ability to "reopen" that channel at a later
// time in a concurrency-safe manner.
//
// The zero StartStopper is ready to use, and begins in a "started" state.
type StartStopper struct {
	stoppedCh chan struct{}
	l         sync.RWMutex
}

// Stop closes the channel returned by stop since the last Start call.
func (s *StartStopper) Stop() {
	s.l.Lock()
	defer s.l.Unlock()
	select {
	default:
		if s.stoppedCh == nil {
			s.stoppedCh = make(chan struct{})
		}
		close(s.stoppedCh)
	case <-s.stoppedCh:
		// No-op; already stopped.
	}
}

// Start replaces the internal channel with a new open one.
// All subsequent calls to Stopped will receive this channel.
func (s *StartStopper) Start() {
	s.l.Lock()
	defer s.l.Unlock()
	if s.stoppedCh == nil {
		s.stoppedCh = make(chan struct{})
	}
	select {
	default:
		// No-op; already started.
	case <-s.stoppedCh:
		s.stoppedCh = make(chan struct{})
	}
}

// Stopped returns a channel that blocks forever until Stop is called on this
// StartStopper.
func (s *StartStopper) Stopped() <-chan struct{} {
	s.l.RLock()
	defer s.l.RUnlock()
	if s.stoppedCh == nil {
		s.stoppedCh = make(chan struct{})
	}
	return s.stoppedCh
}

// IsStopped is a convenience method that returns true if in stopped state (i.e.
// the channel returned from Stopped right now is closed, or true otherwise.
func (s *StartStopper) IsStopped() bool {
	s.l.RLock()
	defer s.l.RUnlock()
	if s == nil {
		return false
	}
	select {
	default:
		return false
	case <-s.stoppedCh:
		return true
	}
}
