package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/comm"
	"io/ioutil"
)

type filesSvc struct{}

var _ api.FilesService = new(filesSvc)

func (h *filesSvc) CopyFile(_ context.Context, partyID int64) error {
	return runWithNotifyArchiveChanged(fmt.Sprintf("копирование файла %d", partyID), func(log comm.Logger, ctx context.Context) error {
		return data.CopyParty(partyID)
	})
}

func (h *filesSvc) DeleteFile(_ context.Context, partyID int64) error {
	return runWithNotifyArchiveChanged(fmt.Sprintf("удаление файла %d", partyID), func(log comm.Logger, ctx context.Context) error {
		return data.DeleteParty(partyID)
	})
}

func (h *filesSvc) SaveFile(_ context.Context, partyID int64, filename string) error {

	var party data.PartyValues
	if err := data.GetPartyValues(partyID, &party, -1); err != nil {
		return err
	}
	b, err := json.MarshalIndent(&party, "", "\t")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, b, 0644); err != nil {
		return err
	}
	return nil
}

func (h *filesSvc) GetCurrentParty(ctx context.Context) (r *apitypes.Party, err error) {
	partyID, err := data.GetCurrentPartyID()
	if err != nil {
		return nil, err
	}
	return h.GetParty(ctx, partyID)
}

func (h *filesSvc) SetCurrentParty(ctx context.Context, partyID int64) error {
	if workgui.IsConnected() {
		return merry.New("нельзя сменить активную партию пока выполняется опрос")
	}
	_, err := data.DB.ExecContext(ctx, `UPDATE app_config SET party_id=? WHERE id=1`, partyID)
	return err
}

func (h *filesSvc) ListParties(ctx context.Context, filterSerial int64) ([]*apitypes.PartyInfo, error) {
	var xs []data.PartyInfo
	const (
		query1 = `
SELECT * FROM party
WHERE exists(SELECT product_id FROM product WHERE product.party_id = party.party_id AND serial = ?)
ORDER BY created_at DESC`
		query2 = ` SELECT * FROM party ORDER BY created_at DESC`
	)
	var err error
	if filterSerial > -1 {
		err = data.DB.SelectContext(ctx, &xs, query1, filterSerial)
	} else {
		err = data.DB.SelectContext(ctx, &xs, query2)
	}

	if err != nil {
		return nil, err
	}

	parties := make([]*apitypes.PartyInfo, 0)
	for _, x := range xs {
		parties = append(parties, &apitypes.PartyInfo{
			PartyID:     x.PartyID,
			Name:        x.Name,
			DeviceType:  x.DeviceType,
			ProductType: x.ProductType,
			CreatedAt:   timeUnixMillis(x.CreatedAt),
		})
	}
	return parties, nil
}

func (h *filesSvc) GetParty(_ context.Context, partyID int64) (*apitypes.Party, error) {
	dataParty, err := data.GetParty(partyID)
	if err != nil {
		return nil, err
	}
	party := &apitypes.Party{
		PartyID:     dataParty.PartyID,
		CreatedAt:   timeUnixMillis(dataParty.CreatedAt),
		Name:        dataParty.Name,
		DeviceType:  dataParty.DeviceType,
		ProductType: dataParty.ProductType,
		Products:    []*apitypes.Product{},
	}

	for _, p := range dataParty.Products {
		party.Products = append(party.Products, convertDataProductToApiProduct(p))
	}
	return party, nil
}

func (h *filesSvc) CreateNewParty(_ context.Context, productsCount int8) error {
	if workgui.IsConnected() {
		return merry.New("нельзя создать новую партию пока выполняется опрос")
	}

	return runWithNotifyArchiveChanged(fmt.Sprintf("создание новой партии: %d приборов", productsCount), func(log comm.Logger, ctx context.Context) error {

		if err := data.SetNewCurrentParty(int(productsCount)); err != nil {
			return err
		}
		party, err := data.GetCurrentPartyInfo()
		if err != nil {
			return err
		}

		d, f := appcfg.DeviceTypes[party.DeviceType]
		if f && d.InitParty != nil {
			if err := d.InitParty(); err != nil {
				return err
			}
		}

		return nil
	})
}

func convertDataProductToApiProduct(p data.Product) *apitypes.Product {
	return &apitypes.Product{
		ProductID:      p.ProductID,
		PartyID:        p.PartyID,
		PartyCreatedAt: timeUnixMillis(p.CreatedAt),
		Comport:        p.Comport,
		Addr:           int8(p.Addr),
		Active:         p.Active,
		Serial:         int64(p.Serial),
	}
}
