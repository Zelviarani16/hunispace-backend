package route

import (
	"github.com/gofiber/fiber/v2"

	"hunispace/app/repository"
	"hunispace/app/service"
)

func init() {
	register(func(api fiber.Router, d Deps) {
		svc := service.NewInvoiceService(
			d.Pool,
			repository.NewLeaseRepository(),
			repository.NewInvoiceRepository(),
		)

		invoices := api.Group("/invoices", d.Auth)
		invoices.Get("/mine", d.Tenant, svc.Mine)
		invoices.Get("", d.Owner, svc.List)
		invoices.Post("", d.Owner, svc.Create)
		invoices.Post("/:id/pay", d.Tenant, svc.Pay)
		invoices.Post("/:id/verify", d.Owner, svc.Verify)
	})
}
