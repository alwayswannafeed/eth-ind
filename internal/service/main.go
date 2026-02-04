package service

import (
	"context"
	"net"
	"net/http"

	"github.com/alwayswannafeed/eth-ind/internal/config"
	"github.com/alwayswannafeed/eth-ind/internal/data"
	"github.com/alwayswannafeed/eth-ind/internal/data/pg"
	"github.com/alwayswannafeed/eth-ind/internal/service/indexer"
	"gitlab.com/distributed_lab/kit/copus/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func Run(ctx context.Context, cfg config.Config) {
	if err := newService(cfg).run(ctx); err != nil {
		panic(err)
	}
}

type service struct {
	log      *logan.Entry
	cfg      config.Config
	copus    types.Copus
	listener net.Listener

	storage  data.Storage     
	indexer  *indexer.Indexer 
}

func newService(cfg config.Config) *service {
	storage := pg.New(cfg.DB())
	idx := indexer.NewIndexer(cfg, storage)

	return &service{
		log:      cfg.Log(),
		cfg:      cfg,
		copus:    cfg.Copus(),
		listener: cfg.Listener(),
		storage:  storage,
		indexer:  idx,
	}
}

func (s *service) run(ctx context.Context) error {
	s.log.Info("Service started")

	go func() {
		if err := s.indexer.Run(ctx); err != nil {
			s.log.WithError(err).Error("Indexer crashed or stopped")
		}
	}()

	r := s.router()

	if err := s.copus.RegisterChi(r); err != nil {
		return errors.Wrap(err, "cop failed")
	}

	s.log.Infof("Listening on %s", s.listener.Addr())
	return http.Serve(s.listener, r)
}