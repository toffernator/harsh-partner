package lamport

import (
   "sync"
)

type LamportClock struct {
   time uint32
   mutex sync.RWMutex
}

func (l *LamportClock) Tick() {
   l.mutex.Lock();
   defer l.mutex.Unlock()

   l.time += 1
}

func (l *LamportClock) TickAgainst(other *LamportClock) {
   l.mutex.Lock()
   defer l.mutex.Unlock()

   l.time = max(l.time, other.time) + 1
}

func (l *LamportClock) Read() uint32 {
   l.mutex.RLock()
   defer l.mutex.RUnlock()

   return l.time
}

func max(a, b uint32) uint32 {
   if a < b {
      return b
   }
   return a
}
