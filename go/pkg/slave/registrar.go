package slave

import (
	"context"
	"crypto/tls"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/dsl/reader"
	pb "github.com/shibukazu/open-ve/go/proto/slave/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type SlaveRegistrar struct {
	Id         string
	Address    string
	TLSEnabled bool
	dslReader  *reader.DSLReader
	gRPCClient pb.SlaveServiceClient
	gRPCConn   *grpc.ClientConn
	logger     *slog.Logger
}

func NewSlaveRegistrar(id, slaveAddress string, slaveTLSEnabled bool, masterAddress string, masterTLSEnabled bool, dslReader *reader.DSLReader, logger *slog.Logger) *SlaveRegistrar {
	var opts []grpc.DialOption

	if masterTLSEnabled {
		creds := credentials.NewTLS(&tls.Config{})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))

	conn, err := grpc.Dial(masterAddress, opts...)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	gRPCClient := pb.NewSlaveServiceClient(conn)

	return &SlaveRegistrar{
		Id:         id,
		Address:    slaveAddress,
		TLSEnabled: slaveTLSEnabled,
		dslReader:  dslReader,
		gRPCClient: gRPCClient,
		gRPCConn:   conn,
		logger:     logger,
	}
}

func (s *SlaveRegistrar) RegisterTimer(ctx context.Context, wg *sync.WaitGroup) {
	s.logger.Info("🟢 slave registration timer started")
	s.register(ctx)
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			s.gRPCConn.Close()
			s.logger.Info("🛑 slave registration timer stopped")
			wg.Done()
			return
		case <-ticker.C:
			s.register(ctx)
		}
	}
}

func (s *SlaveRegistrar) register(ctx context.Context) {
	dsl, err := s.dslReader.Read(ctx)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	validationIds := make([]string, len(dsl.Validations))
	for i, validation := range dsl.Validations {
		validationIds[i] = validation.ID
	}
	_, err = s.gRPCClient.Register(ctx, &pb.RegisterRequest{
		Id:            s.Id,
		Address:       s.Address,
		TlsEnabled:    s.TLSEnabled,
		ValidationIds: validationIds,
	})
	if err != nil {
		s.logger.Error(failure.Translate(err, appError.ErrSlaveRegistrationFailed, failure.Message("Failed to register to master")).Error())
	} else {
		s.logger.Info("📓 slave registration success")
	}
}
