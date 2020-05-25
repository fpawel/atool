package app

import (
	"context"
	"database/sql"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"strconv"
	"strings"
)

type prodPrmSvc struct{}

var _ api.ProductParamService = new(prodPrmSvc)

func (h *prodPrmSvc) SetValue(_ context.Context, key string, productID int64, valueStr string) error {

	if len(strings.TrimSpace(valueStr)) == 0 {
		_, err := data.DB.Exec(`DELETE FROM product_value WHERE product_id = ? AND key = ?`, productID, key)
		return err
	}

	value, err := strconv.ParseFloat(strings.ReplaceAll(valueStr, ",", "."), 64)
	if err != nil {
		return err
	}

	r, err := data.DB.Exec(`
INSERT INTO product_value(product_id, key, value) VALUES (?,?,?)
ON CONFLICT (product_id,key) DO UPDATE SET value = ? `, productID, key, value, value)
	if err != nil {
		return err
	}
	n, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return merry.Errorf("n=%d, expected 1", n)
	}
	return nil
}

func (h *prodPrmSvc) GetValue(_ context.Context, key string, productID int64) (string, error) {
	var v float64
	err := data.DB.Get(&v,
		`SELECT value FROM product_value WHERE product_id = ? AND key = ?`,
		productID, key)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return formatFloat(v), nil
}
