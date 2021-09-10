package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"ivanrybin/work/avito_bank_service/pkg/bank"
	"ivanrybin/work/avito_bank_service/pkg/config"
)

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc

	cfg config.ServerConfig

	bank       bank.Bank
	httpServer *http.Server
}

func New(ctx context.Context, bank bank.Bank, cfg config.ServerConfig) *Server {
	server := &Server{
		bank: bank,
		cfg:  cfg,
	}

	if ctx != nil {
		server.ctx, server.cancel = context.WithCancel(ctx)
	} else {
		server.ctx, server.cancel = context.WithCancel(context.Background())
	}

	router := mux.NewRouter().StrictSlash(false)
	router.HandleFunc("/", server.homeHandler).Methods("GET")
	router.HandleFunc("/add", server.addHandler).Methods("POST")
	router.HandleFunc("/balance", server.balanceHandler).Methods("GET")
	router.HandleFunc("/withdraw", server.withdrawHandler).Methods("POST")
	router.HandleFunc("/transfer", server.transferHandler).Methods("POST")

	server.httpServer = &http.Server{
		Addr:        cfg.Address(),
		Handler:     router,
		ReadTimeout: time.Second * 30,
	}

	return server
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) ShutDown() error {
	defer s.cancel()

	ctx, _ := context.WithTimeout(s.ctx, time.Second*10)

	return s.httpServer.Shutdown(ctx)
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Welcome!"))
}

func (s *Server) balanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req BalanceRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	currency := r.URL.Query().Get("currency")

	if balance, err := s.bank.Balance(req.AccountId, currency); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(ErrorResponse{Error: err.Error()}.Marshal())
	} else {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(BalanceResponse{CentsSum: balance}.Marshal())
	}
}

func (s *Server) addHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req AddRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.bank.Add(req.AccountId, req.CentsSum); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(ErrorResponse{Error: err.Error()}.Marshal())
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) withdrawHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req WithDrawRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.bank.Withdraw(req.AccountId, req.CentsSum); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(ErrorResponse{Error: err.Error()}.Marshal())
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) transferHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req TransferRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.bank.Transfer(req.SenderId, req.RecipientId, req.CentsSum); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(ErrorResponse{Error: err.Error()}.Marshal())
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
