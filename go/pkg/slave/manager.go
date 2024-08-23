package slave

import (
	"log/slog"
	"sync"

	"github.com/morikuni/failure/v2"
)

type SlaveManager struct {
	Slaves map[string]*Slave
	logger *slog.Logger
	mu     sync.RWMutex
}

type Slave struct {
	Id            string
	Addr          string
	TLSEnabled    bool
	ValidationIds []string
}

func NewSlaveManager(logger *slog.Logger) *SlaveManager {
	return &SlaveManager{Slaves: map[string]*Slave{}, logger: logger}
}

func (m *SlaveManager) RegisterSlave(id, addr string, tlsEnabled bool, validationIds []string) {
	m.mu.Lock()
	m.Slaves[id] = &Slave{Id: id, Addr: addr, ValidationIds: validationIds, TLSEnabled: tlsEnabled}
	m.mu.Unlock()
}

func (m *SlaveManager) FindSlave(validationId string) (*Slave, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, slave := range m.Slaves {
		for _, id := range slave.ValidationIds {
			if id == validationId {
				return slave, nil
			}
		}
	}
	return nil, failure.New("slave node that can handle the validation ID is not found")
}
