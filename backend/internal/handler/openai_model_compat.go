package handler

import "github.com/Wei-Shaw/sub2api/internal/service"

func openAIChannelForwardModel(mapping service.ChannelMappingResult, requested string) string {
	if mapping.MappedModel != "" { return mapping.MappedModel }
	return requested
}
