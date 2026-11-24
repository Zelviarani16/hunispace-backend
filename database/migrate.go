package database

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"

	"hunispace/migrations"
)

// Migrate menjalankan seluruh berkas *.sql secara berurutan menurut nama.
// Bila dir kosong, dipakai berkas yang dibenamkan ke binary (go:embed); bila
// dir diisi, berkas dibaca dari folder tersebut. Semua berkas ditulis dengan
// IF NOT EXISTS sehingga aman dijalankan berulang.
func Migrate(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	var fsys fs.FS = migrations.FS
	if dir != "" {
		fsys = os.DirFS(dir)
	}
	files, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, file := range files {
		content, err := fs.ReadFile(fsys, file)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(file), err)
		}
	}
	return nil
}
