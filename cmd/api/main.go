package main

import (
	"github.com/huynguyenanh2000/url-shorterner/internal/db"
	"github.com/huynguyenanh2000/url-shorterner/internal/env"
	"github.com/huynguyenanh2000/url-shorterner/internal/idgen"
	"github.com/huynguyenanh2000/url-shorterner/internal/store"
	"github.com/huynguyenanh2000/url-shorterner/internal/store/cache"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			URL Shorterner API
//	@description	API for URL Shorterner Service.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apiKey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	err := godotenv.Load()
	if err != nil {
		logger.Fatal(err)
	}

	cfg := config{
		addr:      env.GetString("ADDR", ":8080"),
		apiURL:    env.GetString("EXTERNAL_URL", "localhost:8080"),
		machineID: env.GetInt("MACHINE_ID", 1),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "admin:adminpassword@tcp(localhost:3306)/url_shorterner?parseTime=true"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		redisCfg: redisConfig{
			addr:   env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:     env.GetString("REDIS_PW", ""),
			db:     env.GetInt("REDIS_DB", 0),
			enable: env.GetBool("REDIS_ENABLE", false),
		},
		env: env.GetString("ENV", "development"),
	}

	// Database
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("database connection pool established")

	// Cache
	var rdb *redis.Client
	if cfg.redisCfg.enable {
		rdb = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
		logger.Info("redis cache connection established")

		defer rdb.Close()
	}

	snowflakeIDGenerator, err := idgen.NewSnowflakeClient(int64(cfg.machineID))
	if err != nil {
		logger.Fatal(err)
	}
	store := store.NewStorage(db)
	cacheStorage := cache.NewRedisStorage(rdb)

	app := &application{
		config:       cfg,
		store:        store,
		cacheStorage: cacheStorage,
		idGenerator:  snowflakeIDGenerator,
		logger:       logger,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
