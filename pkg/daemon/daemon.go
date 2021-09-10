package daemon

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ivanrybin/work/avito_bank_service/pkg/bank"
	"ivanrybin/work/avito_bank_service/pkg/config"
	"ivanrybin/work/avito_bank_service/pkg/db"
	"ivanrybin/work/avito_bank_service/pkg/server"
)

type Daemon struct {
	ctx    context.Context
	cancel context.CancelFunc

	cfg config.Config

	db     db.BankDB
	bank   bank.Bank
	server *server.Server
}

func New(ctx context.Context, cfg config.Config) (d *Daemon, err error) {
	d = &Daemon{cfg: cfg}

	if ctx != nil {
		d.ctx, d.cancel = context.WithCancel(ctx)
	} else {
		d.ctx, d.cancel = context.WithCancel(context.Background())
	}

	d.db, err = db.New(d.ctx, cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	d.bank, err = bank.New(d.ctx, d.db, cfg.Bank)
	if err != nil {
		log.Fatalf("failed to init bank: %v", err)
	}

	d.server = server.New(d.ctx, d.bank, cfg.Server)

	return d, nil
}

func (d *Daemon) Start() error {
	log.Println("daemon started")

	serverErrC := make(chan error, 1)

	go func() {
		log.Printf("listening %s", d.cfg.Server.Address())
		serverErrC <- d.server.Run()
	}()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-d.ctx.Done():
			log.Println("interrupted by main context")
			return
		case <-stop:
			log.Println("interrupted by syscall")
			d.cancel()
		}
	}()

	select {
	case <-d.ctx.Done():
		d.Stop()
		return nil
	case err := <-serverErrC:
		d.Stop()
		return err
	}
}

func (d *Daemon) Stop() {
	defer d.cancel()

	if err := d.server.ShutDown(); err != nil {
		log.Printf("server closing error: %v", err)
	}

	if err := d.db.Close(); err != nil {
		log.Printf("db closing error: %v", err)
	}

	log.Println("daemon is shut down")
}
