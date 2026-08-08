package shield

import (
	"sage-backend/internal/shared/config"
	"sage-backend/internal/shared/db"
	shield_http "sage-backend/internal/shield/adapters/inbound/http"
	"sage-backend/internal/shield/adapters/outbound/postgres"
	"sage-backend/internal/shield/ports/inbound"
	"sage-backend/internal/shield/usecase"
	"sage-backend/pkg/crypto"

	"github.com/go-resty/resty/v2"
)

type Module struct {
	LogsUseCase        inbound.LogsUseCase
	LogsDataUseCase    inbound.LogsDataUseCase
	DataQualityUseCase inbound.DataQualityUseCase
	ParserUseCase      inbound.ParserUseCase
	IntegrationUseCase inbound.IntegrationUseCase

	EventHandler       *shield_http.EventHandler
	QualityHandler     *shield_http.QualityHandler
	ParserHandler      *shield_http.ParserHandler
	IntegrationHandler *shield_http.IntegrationHandler
	LogsDataHandler    *shield_http.LogsDataHandler
}

func NewModule(
	database *db.DB,
	appConfig *config.APIConfig,
	encryptor crypto.Encryptor,
	restyClient *resty.Client,
) *Module {
	eventRepo := postgres.NewSecurityEventRepository(database)
	dataSourceRepo := postgres.NewDataSourceRepository(database)
	jobRepo := postgres.NewIngestionJobRepository(database)
	qualityRepo := postgres.NewDataQualityRepository(database)
	parserRepo := postgres.NewParserRepository(database)
	integrationRepo := postgres.NewIntegrationRepository(database)
	parsedLogRepo := postgres.NewParsedLogRepository(database)

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

	eventHandler := shield_http.NewEventHandler(logsUseCase)
	qualityHandler := shield_http.NewQualityHandler(dataQualityUseCase)
	parserHandler := shield_http.NewParserHandler(parserUseCase)
	integrationHandler := shield_http.NewIntegrationHandler(integrationUseCase)
	logsDataHandler := shield_http.NewLogsDataHandlerWithService(logsDataUseCase)

	return &Module{
		LogsUseCase:        logsUseCase,
		LogsDataUseCase:    logsDataUseCase,
		DataQualityUseCase: dataQualityUseCase,
		ParserUseCase:      parserUseCase,
		IntegrationUseCase: integrationUseCase,
		EventHandler:       eventHandler,
		QualityHandler:     qualityHandler,
		ParserHandler:      parserHandler,
		IntegrationHandler: integrationHandler,
		LogsDataHandler:    logsDataHandler,
	}
}
