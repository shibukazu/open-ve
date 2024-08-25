package slave

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/dsl/reader"
)

type SlaveRegistrar struct {
	Id                string
	SlaveHTTPAddress  string
	SlaveTLSEnabled   bool
	MasterHTTPAddress string
	dslReader         *reader.DSLReader
	httpClient        *http.Client
	logger            *slog.Logger
}

func NewSlaveRegistrar(id, slaveHTTPAddress string, slaveTLSEnabled bool, masterHTTPAddress string, dslReader *reader.DSLReader, logger *slog.Logger) *SlaveRegistrar {
	var client *http.Client
	masterTLSEnabled := strings.HasPrefix(masterHTTPAddress, "https")
	if masterTLSEnabled {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{},
		}
		client = &http.Client{Transport: transport}
	} else {
		client = &http.Client{}
	}
	client.Timeout = 5 * time.Second

	return &SlaveRegistrar{
		Id:                id,
		SlaveHTTPAddress:  slaveHTTPAddress,
		SlaveTLSEnabled:   slaveTLSEnabled,
		MasterHTTPAddress: masterHTTPAddress,
		dslReader:         dslReader,
		httpClient:        client,
		logger:            logger,
	}
}

func (s *SlaveRegistrar) RegisterTimer(ctx context.Context, wg *sync.WaitGroup) {
	s.logger.Info("🟢 slave registration timer started")
	s.Register(ctx)
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			s.logger.Info("🛑 slave registration timer stopped")
			wg.Done()
			return
		case <-ticker.C:
			s.Register(ctx)
		}
	}
}

func (s *SlaveRegistrar) Register(ctx context.Context) {
	dsl, err := s.dslReader.Read(ctx)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	validationIds := make([]string, len(dsl.Validations))
	for i, validation := range dsl.Validations {
		validationIds[i] = validation.ID
	}
	reqBody := map[string]interface{}{
		"id":             s.Id,
		"address":        s.SlaveHTTPAddress,
		"tls_enabled":    s.SlaveTLSEnabled,
		"validation_ids": validationIds,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		s.logger.Error(failure.Translate(err, appError.ErrSlaveRegistrationFailed, failure.Message("Failed to marshal request body")).Error())
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.MasterHTTPAddress+"/v1/slave/register", bytes.NewReader(body))
	if err != nil {
		s.logger.Error(failure.Translate(err, appError.ErrSlaveRegistrationFailed, failure.Message("Failed to create request")).Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(failure.Translate(err, appError.ErrSlaveRegistrationFailed, failure.Message("Failed to send request")).Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error(failure.New(appError.ErrSlaveRegistrationFailed, failure.Messagef("Failed to register to master: %d", resp.StatusCode)).Error())
		return
	} else {
		s.logger.Info("📓 slave registration success")
	}
}
