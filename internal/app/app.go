package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/example/prreview/internal/config"
	"github.com/example/prreview/internal/handlers"
	"github.com/example/prreview/internal/repo"
	"github.com/example/prreview/internal/server"
	"github.com/example/prreview/internal/services"
)

type App struct {
	DB     *sqlx.DB
	Router *server.RouterHolder
	Logger *log.Logger
	Repos  *repo.SQLRepo
	Svcs   *services.Services
}
func NewApp(cfg config.Config, logger *log.Logger) (*App, error) {
    db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
    if err != nil {
        return nil, err
    }

    if err := repo.RunMigrations(db); err != nil {
        panic(fmt.Sprintf("migration failed: %v", err))
    }

    repos := repo.NewSQLRepo(db)
    svcs := services.NewServices(repos)
    router := server.NewRouter()

    handlers.RegisterTeamRoutes(router.Mux(), repos, svcs)
    handlers.RegisterUserRoutes(router.Mux(), repos, svcs)
    handlers.RegisterPRRoutes(router.Mux(), repos, svcs)

    router.Mux().PathPrefix("/docs/").Handler(
        http.StripPrefix("/docs/", http.FileServer(http.Dir("/app/swagger-ui"))),
    )

    return &App{
        DB:     db,
        Router: router,
        Logger: logger,
        Repos:  repos,
        Svcs:   svcs,
    }, nil
}
