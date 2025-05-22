package workflow

import (
	"errors"
	"testing"
	"time"

	"github.com/Hiroshi0900/saga-example/activity"
	"github.com/Hiroshi0900/saga-example/compensation"
	"github.com/Hiroshi0900/saga-example/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

// SagaTestSuite はTemporalのテストスイートを拡張したテストスイート
type SagaTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

// TestSagaWorkflowSuite はテストスイートを実行する関数
func TestSagaWorkflowSuite(t *testing.T) {
	suite.Run(t, new(SagaTestSuite))
}

// TestSagaWorkflowCompensation はSagaワークフローの補償処理をテストする
func (s *SagaTestSuite) TestSagaWorkflowCompensation() {
	// given
	// テスト環境のセットアップ
	env := s.NewTestWorkflowEnvironment()

	// テスト用データの作成
	testData := &model.TransactionData{
		ID:        "test-tx-123",
		Status:    "STARTED",
		CreatedAt: time.Now(),
	}

	// アクティビティAのモック（成功）
	// 重要: mock.Anythingを使用して任意の型の引数を受け入れるようにする
	env.OnActivity(activity.ActivityA, mock.Anything, mock.Anything).Return(
		&model.TransactionData{
			ID:        testData.ID,
			Status:    "A_COMPLETED",
			StepAData: "Step A completed",
			CreatedAt: testData.CreatedAt, // 同じCreatedAtを使用
		},
		nil,
	)

	// アクティビティBのモック（成功）
	env.OnActivity(activity.ActivityB, mock.Anything, mock.Anything).Return(
		&model.TransactionData{
			ID:        testData.ID,
			Status:    "B_COMPLETED",
			StepAData: "Step A completed",
			StepBData: "Step B completed",
			CreatedAt: testData.CreatedAt,
		},
		nil,
	)

	// アクティビティCのモック（失敗）
	env.OnActivity(activity.ActivityC, mock.Anything, mock.Anything).Return(
		nil,
		errors.New("Activity C failed as expected"),
	)

	// 補償アクションAのモック（成功）
	env.OnActivity(compensation.CompensationA, mock.Anything, mock.Anything).Return(
		&model.TransactionData{
			ID:        testData.ID,
			Status:    "A_COMPENSATED",
			StepAData: "Step A compensated",
			StepBData: "Step B completed",
			CreatedAt: testData.CreatedAt,
		},
		nil,
	)

	// 補償アクションBのモック（2回失敗後に成功）
	// 3回目までは失敗
	env.OnActivity(compensation.CompensationB, mock.Anything, mock.Anything).
		Return(nil, errors.New("Compensation B failed deliberately")).Times(3)

	// 4回目は成功
	env.OnActivity(compensation.CompensationB, mock.Anything, mock.Anything).
		Return(
			&model.TransactionData{
				ID:        testData.ID,
				Status:    "B_COMPENSATED",
				StepAData: "Step A compensated",
				StepBData: "Step B compensated",
				CreatedAt: testData.CreatedAt,
			},
			nil,
		).Once()

	// when
	// ワークフローの実行
	env.ExecuteWorkflow(SagaWorkflow, testData)

	// then
	// ワークフローが完了していることを確認
	s.True(env.IsWorkflowCompleted())

	// ワークフローがエラーで終了していることを確認
	s.Error(env.GetWorkflowError())

	// エラーメッセージにアクティビティCのエラーメッセージが含まれていることを確認
	s.Contains(env.GetWorkflowError().Error(), "Activity C failed as expected")

	// アクティビティと補償アクションが期待通りの順序で呼び出されたことを確認
	env.AssertExpectations(s.T())
}
