package activity

import (
	"context"
	"fmt"
	"github.com/Hiroshi0900/saga-example/model"
)

func ActivityB(_ context.Context, data *model.TransactionData) (*model.TransactionData, error) {
	fmt.Printf("ActivityB ID: %v\n", data.ID)

	// 本来の実装はまた今度
	data.StepBData = "StepB Completed"
	data.Status = "B_Completed"

	return data, nil
}
