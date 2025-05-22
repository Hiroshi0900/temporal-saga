package compensation

import (
	"context"
	"fmt"
	"github.com/Hiroshi0900/saga-example/model"
)

func CompensationB(ctx context.Context, data *model.TransactionData) (*model.TransactionData, error) {
	fmt.Printf("CompensationB ID: %v\n", data.ID)

	data.StepAData = "Step B Compensated"
	data.Status = "B_Compensated"

	return data, nil
}
