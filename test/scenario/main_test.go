package scenario_test

import (
	"context"
	"log/slog"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Suite")
}

var testCtx context.Context
var cancelFn context.CancelFunc
var postgreSQLContainer *testcontainers.DockerContainer
var _ = BeforeSuite(func() {
	ctx, cancel := context.WithCancel(context.Background())
	testCtx = ctx
	cancelFn = cancel

	logger := slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	logger.InfoContext(ctx, "creating PostgreSQL container...")
	pc, err := createPostgreSQLContainer(ctx)
	Expect(err).Should(Succeed())
	postgreSQLContainer = pc
})

var _ = AfterSuite(func() {
	err := postgreSQLContainer.Terminate(testCtx)
	Expect(err).Should(Succeed())
	cancelFn()

})
