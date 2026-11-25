-- Penolakan bukti bayar oleh owner (Anggota 3). Idempotent.
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS rejection_note TEXT;
