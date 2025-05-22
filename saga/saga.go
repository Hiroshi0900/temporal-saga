package saga

import (
	"fmt"
	"go.temporal.io/sdk/workflow"
)

type (
	// Saga は補償トランザクションをサポートするワークフローの実装
	Saga struct {
		ctx                  workflow.Context
		compensations        []compensationFn
		parallelCompensation bool
		continueWithError    bool
	}

	// Options はSagaのオプション
	Options struct {
		ParallelCompensation bool
		ContinueWithError    bool
	}

	compensationFn func(workflow.Context) error
)

// NewSaga は新しいSagaインスタンスを作成します
func NewSaga(ctx workflow.Context, options Options) *Saga {
	return &Saga{
		ctx:                  ctx,
		compensations:        []compensationFn{},
		parallelCompensation: options.ParallelCompensation,
		continueWithError:    options.ContinueWithError,
	}
}

// AddCompensation は補償アクションを追加します
func (s *Saga) AddCompensation(compensationFn compensationFn) {
	s.compensations = append(s.compensations, compensationFn)
}

// Compensate は全ての補償アクションを逆順に実行します
func (s *Saga) Compensate(ctx workflow.Context) error {
	if ctx == nil {
		ctx = s.ctx
	}

	logger := workflow.GetLogger(ctx)
	logger.Info("Executing compensation logic")

	var err error
	for i := len(s.compensations) - 1; i >= 0; i-- {
		compensationFn := s.compensations[i]

		// 並列実行が有効な場合、Future を使用して非同期実行
		if s.parallelCompensation {
			future := workflow.ExecuteActivity(ctx, compensationFn)
			err = futureError(future, err, s.continueWithError)
		} else {
			// 順次実行の場合
			if e := compensationFn(ctx); e != nil {
				logger.Error("Compensation failed", "Error", e)
				if !s.continueWithError {
					return e
				}
				err = e
			}
		}
	}

	return err
}

// futureError は並列実行のエラーハンドリングを行います
func futureError(f workflow.Future, prevErr error, continueWithError bool) error {
	var err error
	if e := f.Get(nil, nil); e != nil {
		if prevErr == nil || !continueWithError {
			return e
		}
		err = fmt.Errorf("multiple compensation errors: %v, %v", prevErr, e)
	}
	return err
}

// ExecActivity はアクティビティを実行し、補償アクションを登録するヘルパー関数
func (s *Saga) ExecActivity(ctx workflow.Context, activity interface{}, args ...interface{}) workflow.Future {
	return workflow.ExecuteActivity(ctx, activity, args...)
}
