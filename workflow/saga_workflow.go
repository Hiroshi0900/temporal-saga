package workflow

import (
	"github.com/Hiroshi0900/saga-example/activity"
	"github.com/Hiroshi0900/saga-example/compensation"
	"github.com/Hiroshi0900/saga-example/model"
	"github.com/Hiroshi0900/saga-example/saga"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"time"
)

func SagaWorkflow(ctx workflow.Context, data *model.TransactionData) (*model.TransactionData, error) {
	// Logger set up
	logger := workflow.GetLogger(ctx)
	logger.Info("Saga Workflow started", "transactionID", data.ID)

	retryPolicy := &temporal.RetryPolicy{
		MaximumAttempts: 1,
	}

	// アクティビティのオプションを設定
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         retryPolicy,
	}

	corp := temporal.RetryPolicy{
		MaximumAttempts:    10,               // 最大試行回数
		InitialInterval:    1 * time.Second,  // 初回の待機時間
		MaximumInterval:    10 * time.Second, // 最大待機時間
		BackoffCoefficient: 2.0,              // 待機時間の増加係数
	}
	// 補償アクションのオプションを設定
	co := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         &corp,
	}

	// アクティビティ用コンテキスト
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// 補償アクション用コンテキスト
	compensationCtx := workflow.WithActivityOptions(ctx, co)

	// Sagaを初期化
	sg := saga.NewSaga(ctx, saga.Options{
		ParallelCompensation: false, // 補償アクションを順次実行
		ContinueWithError:    true,  // 一部の補償が失敗しても続行
	})

	// アクティビティAを実行
	future := workflow.ExecuteActivity(activityCtx, activity.ActivityA, data)

	var activityAResult *model.TransactionData
	err := future.Get(ctx, &activityAResult)
	if err != nil {
		logger.Error("ActivityA failed", "error", err)
		return nil, err
	}
	data = activityAResult

	// アクティビティAの補償アクションを追加
	sg.AddCompensation(func(ctx workflow.Context) error {
		future := workflow.ExecuteActivity(compensationCtx, compensation.CompensationA, data)

		var compensationResult *model.TransactionData
		err := future.Get(ctx, &compensationResult)
		if err != nil {
			logger.Error("CompensationA failed", "error", err)

			return err
		}
		data = compensationResult
		return nil
	})

	// アクティビティBを実行
	future = workflow.ExecuteActivity(activityCtx, activity.ActivityB, data)
	var activityBResult *model.TransactionData
	err = future.Get(ctx, &activityBResult)
	if err != nil {
		logger.Error("ActivityB failed", "error", err)

		// Bが失敗した場合の補償処理（Aの補償を実行）
		compensationErr := sg.Compensate(ctx)
		if compensationErr != nil {
			logger.Error("Compensation failed after ActivityB failure", "error", compensationErr)
			return nil, compensationErr
		}

		return nil, err
	}
	data = activityBResult

	// アクティビティBの補償アクションを追加
	sg.AddCompensation(func(ctx workflow.Context) error {
		future := workflow.ExecuteActivity(compensationCtx, compensation.CompensationB, data)

		var compensationResult *model.TransactionData
		err := future.Get(ctx, &compensationResult)
		if err != nil {
			logger.Error("CompensationB failed", "error", err)

			return err
		}
		data = compensationResult
		return nil
	})

	// アクティビティCを実行
	future = workflow.ExecuteActivity(activityCtx, activity.ActivityC, data)
	var activityCResult *model.TransactionData
	err = future.Get(ctx, &activityCResult)
	if err != nil {
		logger.Error("ActivityC failed", "error", err)

		// Cの補償アクションを実行
		compensationErr := sg.Compensate(ctx)
		if compensationErr != nil {
			logger.Error("Compensation failed", "error", compensationErr)
			return nil, compensationErr
		}
		return data, err // 元のエラーを返す
	}

	return data, nil
}
