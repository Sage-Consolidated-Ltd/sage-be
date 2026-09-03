package shield

import (
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	"sage-backend/internal/shared/storage/s3"
	shield_http "sage-backend/internal/shield/adapters/inbound/http"
	"sage-backend/internal/shield/adapters/outbound/memory"
	"sage-backend/internal/shield/adapters/outbound/postgres"
	shield_redis "sage-backend/internal/shield/adapters/outbound/redis"
	"sage-backend/internal/shield/ports/inbound"
	"sage-backend/internal/shield/ports/outbound"
	"sage-backend/internal/shield/usecase"
	"sage-backend/pkg/crypto"

	"github.com/go-resty/resty/v2"
	redis_driver "github.com/redis/go-redis/v9"
)

type Module struct {
	LogsUseCase        inbound.LogsUseCase
	LogsDataUseCase    inbound.LogsDataUseCase
	DataQualityUseCase inbound.DataQualityUseCase
	ParserUseCase      inbound.ParserUseCase
	IntegrationUseCase inbound.IntegrationUseCase
	DashboardUseCase   inbound.DashboardUseCase
	UploadUseCase      inbound.UploadUseCase
	QualityEngine      inbound.DataQualityEngine
	IncidentEngine     inbound.IncidentEngine

	EventHandler       *shield_http.EventHandler
	QualityHandler     *shield_http.QualityHandler
	ParserHandler      *shield_http.ParserHandler
	IntegrationHandler *shield_http.IntegrationHandler
	LogsDataHandler    *shield_http.LogsDataHandler
	DashboardHandler   *shield_http.DashboardHandler
	UploadHandler      *shield_http.UploadHandler
}

func NewModule(
	database *db.DB,
	appConfig *config.APIConfig,
	encryptor crypto.Encryptor,
	restyClient *resty.Client,
) *Module {
	return NewModuleWithRedis(database, appConfig, encryptor, restyClient, nil)
}

func NewModuleWithRedis(
	database *db.DB,
	appConfig *config.APIConfig,
	encryptor crypto.Encryptor,
	restyClient *resty.Client,
	redisClient *redis_driver.Client,
) *Module {
	return NewModuleWithServices(database, appConfig, encryptor, restyClient, redisClient, nil, nil)
}

func NewModuleWithServices(
	database *db.DB,
	appConfig *config.APIConfig,
	encryptor crypto.Encryptor,
	restyClient *resty.Client,
	redisClient *redis_driver.Client,
	uploader *s3.Uploader,
	taskPublisher outbound.TaskPublisherInt,
) *Module {
	eventRepo := postgres.NewSecurityEventRepository(database)
	dataSourceRepo := postgres.NewDataSourceRepository(database)
	jobRepo := postgres.NewIngestionJobRepository(database)
	qualityRepo := postgres.NewDataQualityRepository(database)
	parserRepo := postgres.NewParserRepository(database)
	integrationRepo := postgres.NewIntegrationRepository(database)
	parsedLogRepo := postgres.NewParsedLogRepository(database)
	dashboardRepo := postgres.NewDashboardRepository(database)
	logUploadRepo := postgres.NewLogUploadRepository(database)

	var correlationStore outbound.CorrelationStore
	if redisClient != nil {
		correlationStore = shield_redis.NewCorrelationStore(redisClient)
	} else {
		correlationStore = memory.NewCorrelationStore()
	}

	dataQualityEngine := usecase.NewDataQualityEngineWithRedis(redisClient)
	incidentEngine := usecase.NewIncidentEngine(correlationStore)
	_ = usecase.RegisterDefaultWindowsRules(incidentEngine, correlationStore)

	logsUseCase := usecase.NewLogsService(
		eventRepo,
		dataSourceRepo,
		jobRepo,
		parsedLogRepo,
	)

	logsDataUseCase := usecase.NewLogsDataService(
		dataSourceRepo,
		eventRepo,
		jobRepo,
		nil,
	)

	dataQualityUseCase := usecase.NewDataQualityService(
		qualityRepo,
		parserRepo,
		dataSourceRepo,
		jobRepo,
		taskPublisher,
	)

	parserUseCase := usecase.NewParserService(
		parserRepo,
		eventRepo,
		dataSourceRepo,
		jobRepo,
	)

	integrationUseCase := usecase.NewDataSourceService(
		dataSourceRepo,
		integrationRepo,
		encryptor,
		restyClient,
	)

	dashboardUseCase := usecase.NewDashboardService(dashboardRepo)

	uploadUseCase := usecase.NewUploadService(
		uploader,
		logUploadRepo,
		taskPublisher,
		dataSourceRepo,
	)

	eventHandler := shield_http.NewEventHandler(logsUseCase)
	qualityHandler := shield_http.NewQualityHandler(dataQualityUseCase)
	parserHandler := shield_http.NewParserHandler(parserUseCase)
	integrationHandler := shield_http.NewIntegrationHandler(integrationUseCase)
	logsDataHandler := shield_http.NewLogsDataHandlerWithService(logsDataUseCase)
	dashboardHandler := shield_http.NewDashboardHandler(dashboardUseCase)
	uploadHandler := shield_http.NewUploadHandler(uploadUseCase)

	return &Module{
		LogsUseCase:        logsUseCase,
		LogsDataUseCase:    logsDataUseCase,
		DataQualityUseCase: dataQualityUseCase,
		ParserUseCase:      parserUseCase,
		IntegrationUseCase: integrationUseCase,
		DashboardUseCase:   dashboardUseCase,
		UploadUseCase:      uploadUseCase,
		QualityEngine:      dataQualityEngine,
		IncidentEngine:     incidentEngine,
		EventHandler:       eventHandler,
		QualityHandler:     qualityHandler,
		ParserHandler:      parserHandler,
		IntegrationHandler: integrationHandler,
		LogsDataHandler:    logsDataHandler,
		DashboardHandler:   dashboardHandler,
		UploadHandler:      uploadHandler,
	}
}
