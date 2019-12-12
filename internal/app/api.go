package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/gui"
	"github.com/fpawel/atool/internal/cfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/pkg/must"
	"github.com/fpawel/atool/internal/pkg/winapi"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/lxn/win"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type mainSvc struct{}

var _ api.MainService = new(mainSvc)

func (h *mainSvc) OpenGuiClient(_ context.Context, hWnd int64) error {
	gui.SetHWndTargetSendMessage(win.HWND(hWnd))
	return nil
}

func (h *mainSvc) CloseGuiClient(_ context.Context) error {
	gui.SetHWndTargetSendMessage(win.HWND_TOP)
	return nil
}

func (h *mainSvc) ListParamAddresses(_ context.Context) (r []int32, _ error) {
	for _, n := range cfg.Get().Hardware.ParamAddresses() {
		r = append(r, int32(n))
	}
	return
}

func (h *mainSvc) EditConfig(ctx context.Context) error {

	filename := filepath.Join(tmpDir, "config.yaml")

	if err := ioutil.WriteFile(filename, must.MarshalYaml(cfg.Get()), 0644); err != nil {
		return err
	}
	cmd := exec.Command("./npp/notepad++.exe", filename)
	if err := cmd.Start(); err != nil {
		return err
	}
	winapi.ActivateWindowByPid(cmd.Process.Pid)

	applyConfig := func() error {
		if err := cmd.Wait(); err != nil {
			return err
		}
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		return cfg.SetYaml(b)
	}

	go func() {
		if err := applyConfig(); err != nil {
			log.PrintErr(err)
			go gui.PopupError(merry.Append(err, "Ошибка при сохранении конфигурации"))
			return
		}
		gui.NotifyCurrentPartyChanged()
	}()
	return nil
}

func (h *mainSvc) GetCurrentParty(ctx context.Context) (r *apitypes.Party, err error) {
	partyID, err := data.GetCurrentPartyID(db)
	if err != nil {
		return nil, err
	}
	return h.GetParty(ctx, partyID)
}

func (h *mainSvc) SetCurrentParty(ctx context.Context, partyID int64) error {
	if connected() {
		return merry.New("нельзя сменить активную партию пока выполняется опрос")
	}
	_, err := db.ExecContext(ctx, `UPDATE app_config SET party_id=? WHERE id=1`, partyID)
	return err
}

func (h *mainSvc) ListParties(ctx context.Context) (parties []*apitypes.PartyInfo, err error) {
	var xs []data.PartyInfo
	if err = db.SelectContext(ctx, &xs, `SELECT * FROM party ORDER BY created_at`); err != nil {
		return
	}
	for _, x := range xs {
		parties = append(parties, &apitypes.PartyInfo{
			PartyID:   x.PartyID,
			CreatedAt: timeUnixMillis(x.CreatedAt),
			Name:      x.Name,
		})
	}
	return
}

func (h *mainSvc) GetParty(_ context.Context, partyID int64) (*apitypes.Party, error) {
	dataParty, err := data.GetParty(db, partyID)
	if err != nil {
		return nil, err
	}
	party := &apitypes.Party{
		PartyID:   dataParty.PartyID,
		CreatedAt: timeUnixMillis(dataParty.CreatedAt),
		Name:      dataParty.Name,
		Products:  []*apitypes.Product{},
	}

	for _, p := range dataParty.Products {
		party.Products = append(party.Products, &apitypes.Product{
			ProductID:      p.ProductID,
			PartyID:        p.PartyID,
			PartyCreatedAt: timeUnixMillis(p.PartyCreatedAt),
			Comport:        p.Comport,
			Addr:           int8(p.Addr),
			Device:         p.Device,
			Active:         p.Active,
			Serial:         int64(p.Serial),
		})
	}
	return party, nil
}

func formatIDs(ids []int64) string {
	var ss []string
	for _, id := range ids {
		ss = append(ss, strconv.FormatInt(id, 10))
	}
	return strings.Join(ss, ",")
}

func (h *mainSvc) ListDevices(ctx context.Context) (xs []string, err error) {
	for _, d := range cfg.Get().Hardware {
		xs = append(xs, d.Name)
	}
	return
}

func (h *mainSvc) CreateNewParty(ctx context.Context, productsCount int8, name string) error {
	if connected() {
		return merry.New("нельзя создать новую партию пока выполняется опрос")
	}
	return data.CreateNewParty(ctx, db, int(productsCount), name)
}

func timeUnixMillis(t time.Time) apitypes.TimeUnixMillis {

	return apitypes.TimeUnixMillis(t.UnixNano() / int64(time.Millisecond))
}

func unixMillisToTime(m apitypes.TimeUnixMillis) time.Time {
	t := int64(time.Millisecond) * int64(m)
	sec := t / int64(time.Second)
	ns := t % int64(time.Second)
	return time.Unix(sec, ns)
}
