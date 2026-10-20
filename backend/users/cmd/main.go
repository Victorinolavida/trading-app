package main

import (
	"context"
	"errors"
	"log/slog"
	conf "pkg/config"
	"pkg/database"
	"pkg/logger"

	"users/config"
	"users/internal/adapters/secondary_adapter/postgres"
	"users/internal/domain"
	"users/internal/domainerrors"
)

func main() {
	log := logger.NewLogger()

	conf, err := conf.NewConfig[config.Config]()
	if err != nil {
		logger.Fatal("error loading configs", slog.Any("error", err))
	}

	pool, err := database.NewPool(context.Background(), conf.Database)
	if err != nil {
		logger.Fatal("error connecting to db", slog.Any("error", err))
	}

	// user repo
	repo := postgres.NewUserRepository(pool)
	// TODO: only for test remove
	u, err := domain.NewUser("victorino@test.com", "victorino", "lavida", "super_duper_pass")
	if err != nil {
		logger.Fatal("error creating user", slog.Any("error", err))
	}
	email, err := domain.NewEmail("victorino@test.com")
	if err != nil {
		logger.Fatal("error creating email", slog.Any("error", err))
	}
	exist, err := repo.ExistByEmail(context.Background(), email)
	if err != nil {
		logger.Fatal("error searching for user", slog.Any("error", err))
	}
	if !exist {
		log.Warn("user does not exist, saving it")
		err = repo.Save(context.Background(), u)
		if err != nil {
			logger.Fatal("error saving user", slog.Any("error", err))
		}
	}
	user, err := repo.GetByEmail(context.Background(), email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			logger.Fatal("user not found", slog.Any("error", err))
		}
		logger.Fatal("error retrieving user by email", slog.Any("error", err))
	}
	log.Info("user found!", slog.String("email", user.Email().String()), slog.String("id", user.ID().String()))
	log.Info("check user password")

	pass := "super_duper_pass"
	if domain.VerifyPassword(pass, string(user.Hash())) {
		log.Info("password match")
	} else {
		log.Warn("password does not match")
	}
}
