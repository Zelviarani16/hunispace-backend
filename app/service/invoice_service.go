package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"hunispace/app/model"
	"hunispace/app/repository"
	"hunispace/helper"
)

type InvoiceService struct {
	pool     *pgxpool.Pool
	leases   repository.LeaseRepository
	invoices repository.InvoiceRepository
}

func NewInvoiceService(
	pool *pgxpool.Pool,
	leases repository.LeaseRepository,
	invoices repository.InvoiceRepository,
) *InvoiceService {
	return &InvoiceService{pool: pool, leases: leases, invoices: invoices}
}

func (s *InvoiceService) Create(c *fiber.Ctx) error {
	var req model.InvoiceCreateRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return err
	}
	due, fields := validateInvoiceRequest(req)
	if len(fields) > 0 {
		return helper.Validation(fields)
	}

	ctx := c.UserContext()
	lease, err := s.leases.FindByID(ctx, s.pool, req.LeaseID)
	if err != nil {
		return lookupError(err, "kontrak")
	}
	if lease.Status != model.LeaseActive {
		return helper.Conflict("LEASE_NOT_ACTIVE", "tagihan hanya dapat dibuat untuk kontrak aktif")
	}

	suffix, err := helper.RandomHex(3)
	if err != nil {
		return helper.Internal(err)
	}

	invoice, err := s.invoices.Create(ctx, s.pool, model.Invoice{
		LeaseID:       req.LeaseID,
		InvoiceNumber: buildInvoiceNumber(req.BillingPeriod, req.LeaseID, suffix),
		BillingPeriod: req.BillingPeriod,
		Amount:        req.Amount,
		DueDate:       due,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Conflict("INVOICE_ALREADY_EXISTS", "tagihan untuk periode ini sudah ada")
		}
		return helper.Internal(err)
	}
	return helper.Created(c, "tagihan berhasil dibuat", invoice)
}

func (s *InvoiceService) List(c *fiber.Ctx) error {
	status := c.Query("status")
	switch status {
	case "", model.InvoiceUnpaid, model.InvoicePendingVerification, model.InvoicePaid, model.InvoiceOverdue:
	default:
		return helper.BadRequest("status harus unpaid, pending_verification, paid, atau overdue")
	}
	invoices, err := s.invoices.FindAll(c.UserContext(), s.pool, status)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar tagihan berhasil diambil", invoices)
}

func (s *InvoiceService) Mine(c *fiber.Ctx) error {
	auth := helper.CurrentUser(c)
	invoices, err := s.invoices.FindByTenant(c.UserContext(), s.pool, auth.UserID)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "tagihan saya berhasil diambil", invoices)
}

func (s *InvoiceService) Pay(c *fiber.Ctx) error {
	id, err := helper.ParseID(c, "id")
	if err != nil {
		return err
	}
	var req model.InvoicePayRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return err
	}
	if fields := validateInvoicePay(req); len(fields) > 0 {
		return helper.Validation(fields)
	}

	auth := helper.CurrentUser(c)
	ok, err := s.invoices.MarkPendingVerification(
		c.UserContext(), s.pool, id, auth.UserID,
		strings.TrimSpace(req.PaymentMethod), strings.TrimSpace(req.ProofURL))
	if err != nil {
		return helper.Internal(err)
	}
	if !ok {
		return helper.Conflict("INVOICE_NOT_PAYABLE",
			"tagihan tidak ditemukan, bukan milik anda, atau sudah dibayar / menunggu verifikasi")
	}

	invoice, err := s.invoices.FindByID(c.UserContext(), s.pool, id)
	if err != nil {
		return lookupError(err, "tagihan")
	}
	return helper.Success(c, fiber.StatusOK, "konfirmasi pembayaran diterima, menunggu verifikasi owner", invoice)
}

func (s *InvoiceService) Verify(c *fiber.Ctx) error {
	id, err := helper.ParseID(c, "id")
	if err != nil {
		return err
	}
	ctx := c.UserContext()

	ok, err := s.invoices.MarkPaid(ctx, s.pool, id)
	if err != nil {
		return helper.Internal(err)
	}
	if !ok {
		return helper.Conflict("INVOICE_NOT_PENDING", "tagihan tidak ditemukan atau tidak sedang menunggu verifikasi")
	}

	invoice, err := s.invoices.FindByID(ctx, s.pool, id)
	if err != nil {
		return lookupError(err, "tagihan")
	}
	return helper.Success(c, fiber.StatusOK, "pembayaran terverifikasi", invoice)
}
