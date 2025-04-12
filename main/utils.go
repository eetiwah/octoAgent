package main

import (
	"encoding/json"

	"cwtch.im/cwtch/model"
)

func packageReply(msg string) string {
	return string(cwtchbot.PackMessage(model.OverlayChat, msg))
}

func packageActionableReply(msg string) string {
	return string(cwtchbot.PackMessage(ActionableMessageOverlay, msg))
}

func Unwrap(onion int, msg string) *OverlayEnvelope {
	var envelope OverlayEnvelope
	err := json.Unmarshal([]byte(msg), &envelope)
	if err != nil {
		//		log.Errorf("json error: %v", err)
		return nil
	}
	envelope.onion = onion
	return &envelope
}

func inList(item string, list []string) bool {
	isInSlice := false
	for _, listItem := range list {
		if listItem == item {
			isInSlice = true
			break
		}
	}
	return isInSlice
}
