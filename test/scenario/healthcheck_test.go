package scenario_test

import (
	"context"
	"log/slog"
	"time"

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
		host, err := postgreSQLContainer.Host(ctx)
		Expect(err).Should(Succeed())

		// Get the dynamically mapped port
		mappedPort, err := postgreSQLContainer.MappedPort(ctx, "5432")
		Expect(err).Should(Succeed())

		pgConfig := pg.Config{
			Host:     host,
			Port:     mappedPort.Int(),
			User:     "admin_api",
			Password: "password",
			DBName:   "tacokumo_admin_db",
			SSLMode:  "disable",
		}
		pgxConfig, err := pgxpool.ParseConfig(pgConfig.DSN())
		Expect(err).Should(Succeed())

		p, err := pgxpool.NewWithConfig(ctx, pgxConfig)
		Expect(err).Should(Succeed())

		Eventually(func(ctx context.Context) error {
			err := p.Ping(ctx)
			return err
		}).WithContext(ctx).WithTimeout(10 * time.Second).WithPolling(1 * time.Second).Should(Succeed())
		pgxPool = p

	})

	AfterEach(func() {
		if pgxPool != nil {
			pgxPool.Close()
		}
	})

	It("should return ok", func(ctx context.Context) {
		queries := admindb.New(pgxPool)
		svc := v1alpha1.NewService(slog.Default(), queries, nil, nil, nil, "", 0)
		resp, err := svc.GetLivenessCheck(ctx)
		Expect(err).Should(Succeed())
		Expect(resp).ShouldNot(BeNil())
		Expect(resp.Status).Should(Equal("ok"))
	})
})
