package ui

import "redscout/models"

type StateUpdateMsg struct {
	State *models.State
}

type ScanCompleteMsg struct {
	State *models.State
}

type ErrorMsg struct {
	Err error
}
