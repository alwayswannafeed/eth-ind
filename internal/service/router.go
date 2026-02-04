package service

import (
	"github.com/alwayswannafeed/eth-ind/internal/service/handlers"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
)

func (s *service) router() chi.Router {
	r := chi.NewRouter()

	r.Use(
		ape.RecoverMiddleware(s.log),
		ape.LoganMiddleware(s.log),
		ape.CtxMiddleware(
			handlers.CtxLog(s.log),
			// 👇 ДОДАЄМО БД В КОНТЕКСТ
			// Тепер у кожному запиті буде доступна база через handlers.Storage(r)
			handlers.CtxStorage(s.storage),
		),
	)

	r.Route("/integrations/eth-ind", func(r chi.Router) {
		r.Get("/status", handlers.GetStatus) //info about earliest block in db
		r.Post("/transfers", handlers.GetTransfers)
		r.Get("/transfer/{id}", handlers.GetTransferByID)
	})

	return r
}