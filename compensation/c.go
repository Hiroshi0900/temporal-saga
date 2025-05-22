package compensation

import (
	"context"
	"fmt"
	"github.com/Hiroshi0900/saga-example/model"
)

func CompensationC(ctx context.Context, data *model.TransactionData) (*model.TransactionData, error) {
	fmt.Printf("CompensationC ID: %v\n", data.ID)

	data.StepAData = "Step C Compensated"
	data.Status = "C_Compensated"

	return data, nil
}
