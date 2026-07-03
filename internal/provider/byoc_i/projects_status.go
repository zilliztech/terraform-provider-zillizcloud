package byoc_op

func isBYOCOpProjectAgentBootstrapRequiredStatus(status int) bool {
	return !isBYOCOpProjectAgentReadyStatus(status)
}
