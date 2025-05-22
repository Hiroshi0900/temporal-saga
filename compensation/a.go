package compensation

import (
	"context"
	"fmt"
	"github.com/Hiroshi0900/saga-example/model"
)

func CompensationA(ctx context.Context, data *model.TransactionData) (*model.TransactionData, error) {
	fmt.Printf("CompensationA ID: %v\n", data.ID)

	data.StepAData = "Step A Compensated"
	data.Status = "A_Compensated"

	return data, nil
}
