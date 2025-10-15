package scenario

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tacokumo/admin-api/pkg/apis/v1alpha1"
	"github.com/tacokumo/admin-api/pkg/db/admindb"
	"github.com/tacokumo/admin-api/pkg/pg"
)

var _ = Describe("v1alpha liveness check", func() {
	var pgxPool *pgxpool.Pool
	BeforeEach(func(ctx context.Context) {
		pgConfig := pg.Config{
			Host:     "localhost",
			Port:     5432,
			User:     "user",
			Password: "password",
			DBName:   "admindb",
			SSLMode:  "disable",
		}
		pgxConfig, err := pgxpool.ParseConfig(pgConfig.DSN())
		Expect(err).Should(Succeed())

		p, err := pgxpool.NewWithConfig(ctx, pgxConfig)
		Expect(err).Should(Succeed())
		pgxPool = p

	})

	AfterEach(func() {
		pgxPool.Close()
	})

	It("should return ok", func(ctx context.Context) {
		queries := admindb.New(pgxPool)
		svc := v1alpha1.NewService(slog.Default(), queries)
		resp, err := svc.GetLivenessCheck(ctx)
		Expect(err).Should(Succeed())
		Expect(resp).ShouldNot(BeNil())
		Expect(resp.Status).Should(Equal("ok"))
	})
})
