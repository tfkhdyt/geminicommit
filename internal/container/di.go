package container

import "github.com/tfkhdyt/geminicommit/internal/delivery/cli/handler"

var rootHandler *handler.RootHandler

func init() {
	rootHandler = handler.NewRootHandler()
}

func GetRootHandlerInstance() *handler.RootHandler {
	return rootHandler
}
