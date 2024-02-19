package container

import (
	"github.com/tfkhdyt/geminicommit/internal/delivery/cli/handler"
	"github.com/tfkhdyt/geminicommit/internal/service"
	"github.com/tfkhdyt/geminicommit/internal/usecase"
)

var (
	rootHandler   *handler.RootHandler
	rootUsecase   *usecase.RootUsecase
	gitService    *service.GitService
	geminiService *service.GeminiService
)

func init() {
	gitService = service.NewGitService()
	geminiService = service.NewGeminiService()
	rootUsecase = usecase.NewRootUsecase(gitService, geminiService)
	rootHandler = handler.NewRootHandler(rootUsecase)
}

func GetRootHandlerInstance() *handler.RootHandler {
	return rootHandler
}
