package lamport

import (
   "sync"
   "chkg.com/chitty-chat/api"
)

type LamportClock struct {
   api.Lamport
   mutex sync.RWMutex
}

func (l *LamportClock) Tick() {
   l.mutex.Lock();
   defer l.mutex.Unlock()

   l.Lamport.Time += 1
}

func (l *LamportClock) TickAgainst(otherTime uint32) {
   l.mutex.Lock()
   defer l.mutex.Unlock()

   l.Lamport.Time = max(l.Lamport.Time, otherTime) + 1
}

func (l *LamportClock) Read() uint32 {
   l.mutex.RLock()
   defer l.mutex.RUnlock()

   return l.Lamport.Time
}

func max(a, b uint32) uint32 {
   if a < b {
      return b
   }
   return a
}
