package activity

import (
	"context"
	"fmt"
	"github.com/Hiroshi0900/saga-example/model"
)

func ActivityC(_ context.Context, data *model.TransactionData) (*model.TransactionData, error) {
	fmt.Printf("ActivityC ID: %v\n", data.ID)

	// 本来の実装はまた今度
	data.StepCData = "StepC Completed"
	data.Status = "C_Completed"

	return data, nil
}
