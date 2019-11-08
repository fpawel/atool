package app

import (
	"context"
	"github.com/ansel1/merry"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	"github.com/jmoiron/sqlx"
	"time"
)

func (x appConfig) save(db *sqlx.DB, ctx context.Context) error {

	comm.SetEnableLog(x.LogComport)

	for _, x := range x.Hardware {
		if _, err := db.ExecContext(ctx, `
INSERT INTO hardware(device, baud, timeout_get_responses, timeout_end_response, pause, max_attempts_read)
VALUES (?, ?,?,?,?,?)
ON CONFLICT (device) DO UPDATE SET baud=?,
                                   timeout_get_responses=?,
                                   timeout_end_response=?,
                                   pause=?,
                                   max_attempts_read=?`,
			x.Name, x.Baud, x.TimeoutGetResponse, x.TimeoutEndResponse, x.Pause, x.MaxAttemptsRead,
			x.Baud, x.TimeoutGetResponse, x.TimeoutEndResponse, x.Pause, x.MaxAttemptsRead, x.Name); err != nil {
			return log.Err(merry.Append(err, "save config: INSERT INTO hardware ON CONFLICT (device) DO UPDATE"),
				"device", x)
		}
		if _, err := db.ExecContext(ctx, `DELETE FROM param WHERE device = ?`, x.Name); err != nil {
			return log.Err(merry.Append(err, "save config: DELETE FROM param WHERE device"),
				"device", x)
		}
		if len(x.Params) == 0 {
			x.Params = []paramConfig{{
				Var:    0,
				Count:  2,
				Format: "bcd",
			}}
		}
		for _, p := range x.Params {
			if _, err := db.ExecContext(ctx, `
INSERT INTO param(device, var, count, format) 
VALUES (?,?,?,?)`, x.Name, p.Var, p.Count, p.Format); err != nil {
				return log.Err(merry.Append(err, "save config: INSERT INTO param"),
					"device", x, "param", p)
			}
		}
	}

	c, err := openAppConfig(db, ctx)
	if err != nil {
		return log.Err(merry.Append(err, "save config: open previous config"))
	}
lab1:
	for _, c := range c.Hardware {
		for _, d := range x.Hardware {
			if c.Name == d.Name {
				continue lab1
			}
		}
		if c.Name == "default" {
			// нельзя удалять тип прибора с именем default
			continue
		}
		if _, err := db.ExecContext(ctx, `UPDATE product SET device='DEFAULT' WHERE device=?`, c.Name); err != nil {
			err = merry.Append(err, "save config: UPDATE product SET device='DEFAULT' WHERE device")
			return log.Err(err, "device", c)
		}
		if _, err := db.ExecContext(ctx, `DELETE FROM hardware WHERE device=?`, c.Name); err != nil {
			err = merry.Append(err, "save config: DELETE FROM hardware WHERE device")
			return log.Err(err, "device", c)
		}
	}
	return nil
}

func openAppConfig(db *sqlx.DB, ctx context.Context) (appConfig, error) {
	var c appConfig

	if err := db.GetContext(ctx, &c.LogComport, `SELECT log_comport FROM app_config`); err != nil {
		return c, merry.Append(err, "get config from db")
	}

	var xs []struct {
		Device             string     `db:"device"`
		Baud               int        `db:"baud"`
		Pause              int64      `db:"pause"`
		TimeoutGetResponse int64      `db:"timeout_get_responses"`
		TimeoutEndResponse int64      `db:"timeout_end_response"`
		MaxAttemptsRead    int        `db:"max_attempts_read"`
		Var                modbus.Var `db:"var"`
		Count              int        `db:"count"`
		Format             string     `db:"format"`
	}

	if err := db.SelectContext(ctx, &xs, `
SELECT device, baud, pause,  
       timeout_get_responses,  
       timeout_end_response,
       var, count, format, max_attempts_read
FROM param INNER JOIN hardware USING (device)`); err != nil {
		return c, merry.Append(err, "get config from db")
	}

	for _, x := range xs {
		d := c.getDeviceByName(x.Device)
		d.TimeoutGetResponse = time.Duration(x.TimeoutGetResponse)
		d.TimeoutEndResponse = time.Duration(x.TimeoutEndResponse)
		d.Pause = time.Duration(x.Pause)
		d.MaxAttemptsRead = x.MaxAttemptsRead
		d.Baud = x.Baud
		d.setParam(x.Var, x.Count, x.Format)
	}
	return c, nil
}

type appConfig struct {
	LogComport bool           `yaml:"log_comport"`
	Hardware   []deviceConfig `yaml:"hardware"`
}

type deviceConfig struct {
	Name               string        `yaml:"name"`
	Baud               int           `yaml:"baud"`
	TimeoutGetResponse time.Duration `yaml:"timeout_get_response"`
	TimeoutEndResponse time.Duration `yaml:"timeout_end_response"`
	MaxAttemptsRead    int           `yaml:"max_attempts_read"`
	Pause              time.Duration `yaml:"pause"`
	Params             []paramConfig `yaml:"params"`
}

type paramConfig struct {
	Var    modbus.Var `yaml:"var"`
	Count  int        `yaml:"count"`
	Format string     `yaml:"format"`
}

func (x *appConfig) getDeviceByName(s string) (c *deviceConfig) {
	for i := range x.Hardware {
		if x.Hardware[i].Name == s {
			return &x.Hardware[i]
		}
	}
	x.Hardware = append(x.Hardware, deviceConfig{Name: s})
	return &x.Hardware[len(x.Hardware)-1]
}

func (x *deviceConfig) setParam(v modbus.Var, count int, format string) {
	for i := range x.Params {
		if x.Params[i].Var == v {
			x.Params[i].Count = count
			x.Params[i].Format = format
			return
		}
	}
	x.Params = append(x.Params, paramConfig{
		Var:    v,
		Count:  count,
		Format: format,
	})
}
