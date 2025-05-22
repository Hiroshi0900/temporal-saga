package model

import "time"

type (
	// TransactionData はアクティビティ間で共有されるデータ
	TransactionData struct {
		ID        string
		Status    string
		StepAData string
		StepBData string
		StepCData string
		CreatedAt time.Time
	}
)
