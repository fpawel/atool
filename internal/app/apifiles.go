package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/gui"
	"github.com/fpawel/atool/internal/guiwork"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
)

type filesSvc struct{}

var _ api.FilesService = new(filesSvc)

func (h *filesSvc) CopyFile(_ context.Context, partyID int64) error {
	go func() {
		if err := data.CopyParty(db, partyID); err != nil {
			guiwork.JournalErr(log, merry.Appendf(err, "копирование файла %d", partyID))
			return
		}
		gui.NotifyCurrentPartyChanged()
	}()
	return nil

}

func (h *filesSvc) DeleteFile(_ context.Context, partyID int64) error {
	go func() {
		currentPartyID, err := data.GetCurrentPartyID(db)
		if err != nil {
			return
		}
		if err := data.DeleteParty(db, partyID); err != nil {
			guiwork.JournalErr(log, merry.Appendf(err, "удаление файла %d", partyID))
			return
		}
		if currentPartyID == partyID {
			gui.NotifyCurrentPartyChanged()
		}
	}()
	return nil
}

func (h *filesSvc) SaveFile(ctx context.Context, partyID int64, filename string) error {
	return nil
}

func (h *filesSvc) GetCurrentParty(ctx context.Context) (r *apitypes.Party, err error) {
	partyID, err := data.GetCurrentPartyID(db)
	if err != nil {
		return nil, err
	}
	return h.GetParty(ctx, partyID)
}

func (h *filesSvc) SetCurrentParty(ctx context.Context, partyID int64) error {
	if guiwork.IsConnected() {
		return merry.New("нельзя сменить активную партию пока выполняется опрос")
	}
	_, err := db.ExecContext(ctx, `UPDATE app_config SET party_id=? WHERE id=1`, partyID)
	return err
}

func (h *filesSvc) ListParties(ctx context.Context) (parties []*apitypes.PartyInfo, err error) {
	var xs []data.PartyInfo
	if err = db.SelectContext(ctx, &xs, `SELECT * FROM party ORDER BY created_at DESC`); err != nil {
		return
	}
	for _, x := range xs {
		parties = append(parties, &apitypes.PartyInfo{
			PartyID:     x.PartyID,
			Name:        x.Name,
			DeviceType:  x.DeviceType,
			ProductType: x.ProductType,
			CreatedAt:   timeUnixMillis(x.CreatedAt),
		})
	}
	return
}

func (h *filesSvc) GetParty(_ context.Context, partyID int64) (*apitypes.Party, error) {
	dataParty, err := data.GetParty(db, partyID)
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

func (h *filesSvc) CreateNewParty(ctx context.Context, productsCount int8) error {
	if guiwork.IsConnected() {
		return merry.New("нельзя создать новую партию пока выполняется опрос")
	}
	return data.SetNewCurrentParty(ctx, db, int(productsCount))
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
