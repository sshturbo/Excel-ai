package main

import (
	"embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"excel-ai/internal/app" // Novo import
	"excel-ai/pkg/logger"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Inicializar logger
	if err := logger.InitializeFromFile("logger-config.json"); err != nil {
		// Se falhar ao carregar config, usar padrão
		logger.InitializeWithDefaults(logger.INFO)
		fmt.Printf("[INIT] Aviso: não foi possível carregar configuração do logger: %v\n", err)
	}

	logger.AppInfo("Aplicação iniciando...")

	// Configurar captura de panic
	defer func() {
		if r := recover(); r != nil {
			logger.AppFatal(fmt.Sprintf("Panic recuperado: %v", r))
		}
	}()

	// Configurar graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.AppInfo(fmt.Sprintf("Sinal recebido: %v, encerrando graciosamente", sig))
		logger.GetLogger().Close()
		os.Exit(0)
	}()

	// Create an instance of app structure
	application := app.NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "excel-ai",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        application.Startup,  // Métodos agora exportados
		OnShutdown:       application.Shutdown, // Métodos agora exportados
		Bind: []interface{}{
			application,
		},
	})

	if err != nil {
		logger.AppError(fmt.Sprintf("Erro ao iniciar aplicação: %v", err))
		logger.GetLogger().Close()
	}

	logger.AppInfo("Aplicação encerrada")
	logger.GetLogger().Close()
}
