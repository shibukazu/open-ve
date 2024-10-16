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
	"github.com/shibukazu/open-ve/go/pkg/config"
	"github.com/shibukazu/open-ve/go/pkg/dsl/reader"
	"github.com/shibukazu/open-ve/go/pkg/logger"
)

type SlaveRegistrar struct {
	Id                string
	SlaveHTTPAddress  string
	SlaveTLSEnabled   bool
	SlaveAuthn        config.AuthnConfig
	MasterHTTPAddress string
	MasterAuthn       config.AuthnConfig
	dslReader         *reader.DSLReader
	httpClient        *http.Client
	logger            *slog.Logger
}

func NewSlaveRegistrar(id, slaveHTTPAddress string, slaveTLSEnabled bool, slaveAuthn config.AuthnConfig, masterHTTPAddress string, masterAuthn config.AuthnConfig, dslReader *reader.DSLReader, logger *slog.Logger) *SlaveRegistrar {
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
		SlaveAuthn:        slaveAuthn,
		MasterHTTPAddress: masterHTTPAddress,
		MasterAuthn:       masterAuthn,
		dslReader:         dslReader,
		httpClient:        client,
		logger:            logger,
	}
}

func (s *SlaveRegistrar) RegisterTimer(ctx context.Context, wg *sync.WaitGroup) {
	s.logger.Info("🟢 slave registration timer started")
	s.Register(ctx)
	err := s.Register(ctx)
	if err != nil {
		logger.LogError(s.logger, err)
	}
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			s.logger.Info("🛑 slave registration timer stopped")
			wg.Done()
			return
		case <-ticker.C:
			err := s.Register(ctx)
			if err != nil {
				logger.LogError(s.logger, err)
			}
		}
	}
}

func (s *SlaveRegistrar) Register(ctx context.Context) error {
	dsl, err := s.dslReader.Read(ctx)
	if err != nil {
		return err
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
		"authn": map[string]interface{}{
			"method": s.SlaveAuthn.Method,
			"preshared": map[string]interface{}{
				"key": s.SlaveAuthn.Preshared.Key,
			},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return failure.Translate(err, appError.ErrServerError, failure.Message("failed to marshal slave registration request"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.MasterHTTPAddress+"/v1/slave/register", bytes.NewReader(body))
	if err != nil {
		return failure.Translate(err, appError.ErrServerError, failure.Message("failed to create slave registration request"))
	}
	req.Header.Set("Content-Type", "application/json")

	switch s.MasterAuthn.Method {
	case "preshared":
		req.Header.Set("Authorization", "Bearer "+s.MasterAuthn.Preshared.Key)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return failure.Translate(err, appError.ErrServerError, failure.Message("failed to send slave registration request"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return failure.New(appError.ErrServerError, failure.Messagef("failed to register to master: %d", resp.StatusCode))
	} else {
		s.logger.Info("📓 slave registration success")
	}

	return nil
}
