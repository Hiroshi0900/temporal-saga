package activity

import (
	"context"
	"fmt"
	"github.com/Hiroshi0900/saga-example/model"
)

func ActivityA(_ context.Context, data *model.TransactionData) (*model.TransactionData, error) {
	fmt.Printf("ActivityA ID: %v\n", data.ID)

	// 本来の実装はまた今度
	data.StepAData = "StepA Completed"
	data.Status = "A_Completed"

	return data, nil
}
