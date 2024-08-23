package slave

import (
	"log/slog"
	"sync"
)

type SlaveManager struct {
	Slaves map[string]*Slave
	logger *slog.Logger
	mu     sync.RWMutex
}

type Slave struct {
	Id            string
	Addr          string
	ValidationIds []string
}

func NewSlaveManager(logger *slog.Logger) *SlaveManager {
	return &SlaveManager{Slaves: map[string]*Slave{}, logger: logger}
}

func (m *SlaveManager) RegisterSlave(id, addr string, validationIds []string) {
	m.mu.Lock()
	m.Slaves[id] = &Slave{Id: id, Addr: addr, ValidationIds: validationIds}
	m.mu.Unlock()
}
